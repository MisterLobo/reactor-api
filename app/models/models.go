package models

import (
	"context"
	"fmt"
	"os"
	"reactor/utils"
	"strings"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/docker/docker/client"
	"github.com/google/uuid"
)

type ConnectionConfig struct {
	// gorm.Model
	ID        string `gorm:"type:uuid;primarykey" json:"id"`
	Name      string `gorm:"uniqueIndex" json:"name"`
	Type      string `json:"socket_type"`
	Address   string `json:"socket_address"`
	IsDefault bool   `json:"is_default"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `gorm:"index" json:"deleted_at"`
}

func (c *ConnectionConfig) BeforeCreate(tx *gorm.DB) error {
	c.ID = uuid.NewString()
	return nil
}

type ConnectionManager struct {
	db         *gorm.DB
	dbFilename string
}

var connectionManager *ConnectionManager
var configManager *utils.ConfigurationManager

func (c *ConnectionConfig) ToString() string {
	return fmt.Sprintf("%s://%s", c.Type, c.Address)
}

func DefaultConnectionManager() *ConnectionManager {
	if connectionManager == nil {
		connectionManager = &ConnectionManager{}
	}
	return connectionManager
}
func (c *ConnectionManager) InitDefaults() {
	c.dbFilename = "data.db"
	if c.db == nil {
		c.db = setupDB()
	}
	db := c.db
	dockerHost := os.Getenv("DOCKER_HOST")
	parts := strings.Split(dockerHost, "://")
	socketType, socketAddr := parts[0], parts[1]
	cc := ConnectionConfig{
		Name:      "default",
		Type:      socketType,
		Address:   socketAddr,
		IsDefault: true,
	}
	db.Where(ConnectionConfig{Name: "default"}).FirstOrCreate(&cc)
}
func (c *ConnectionManager) ListConnections() []ConnectionConfig {
	db := c.db
	var conns []ConnectionConfig
	db.Find(&conns)
	return conns
}
func (c *ConnectionManager) GetConnection(id string) (*ConnectionConfig, bool) {
	db := c.db
	conn := ConnectionConfig{ID: id}
	ra := db.First(&conn)
	return &conn, ra.RowsAffected == 1 && &conn != nil
}
func (c *ConnectionManager) GetDefaultConnection() (*ConnectionConfig, string) {
	db := c.db
	var def ConnectionConfig
	db.Where(ConnectionConfig{IsDefault: true}).First(&def)
	return &def, def.ToString()
}
func (c *ConnectionManager) SetDefaultConnection(id string) string {
	db := c.db
	db.Model(&ConnectionConfig{}).Where(&ConnectionConfig{IsDefault: true}).Update("is_default", false)
	upd := ConnectionConfig{
		ID: id,
	}
	db.Model(&upd).Update("is_default", true)
	// db.Save(&upd)
	return id
}
func (c *ConnectionManager) SaveConnection(p *ConnectionConfig) int64 {
	db := c.db
	db.
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(p)
	return db.RowsAffected
}
func (c *ConnectionManager) UpdateConnection(params *ConnectionConfig) int64 {
	db := c.db
	db.Save(params)
	return db.RowsAffected
}
func (c *ConnectionManager) DeleteConnection(id string) int64 {
	db := c.db
	db.Delete(&ConnectionConfig{ID: id})
	return db.RowsAffected
}
func (c *ConnectionManager) TestConnection(id string) (bool, string) {
	cc, ok := c.GetConnection(id)
	if !ok {
		return false, "No connection found"
	}
	cstr := cc.ToString()

	apiClient, err := client.NewClientWithOpts(client.WithHost(cstr), client.WithAPIVersionNegotiation())
	if err != nil {
		return false, err.Error()
	}
	defer apiClient.Close()
	_, err = apiClient.Ping(context.Background())
	if err != nil {
		return false, err.Error()
	}
	return true, ""
}
func (c *ConnectionManager) TestConnectionString(connStr string) (bool, string) {
	apiClient, err := client.NewClientWithOpts(client.WithHost(connStr), client.WithAPIVersionNegotiation())
	if err != nil {
		return false, err.Error()
	}
	defer apiClient.Close()
	_, err = apiClient.Ping(context.Background())
	if err != nil {
		return false, err.Error()
	}
	return true, ""
}

func setupDB() *gorm.DB {
	configManager = utils.DefaultConfigurationManager()
	dataFilePath := fmt.Sprintf("%s/%s", configManager.GetDataPath(), connectionManager.dbFilename)
	db, err := gorm.Open(sqlite.Open(dataFilePath), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize DB: %s", err))
	}
	db.AutoMigrate(&ConnectionConfig{})
	return db
}
