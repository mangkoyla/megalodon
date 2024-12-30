package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/FoolVPN-ID/Megalodon/common/helper"
	logger "github.com/FoolVPN-ID/Megalodon/log"
	"github.com/FoolVPN-ID/Megalodon/sandbox"
	"github.com/FoolVPN-ID/Megalodon/telegram/bot"
	"github.com/tursodatabase/go-libsql"
)

type databaseStruct struct {
	client           *sql.DB
	dbDir            string
	dbName           string
	dbPath           string
	dbUrl            string
	dbToken          string
	dbConnector      *libsql.Connector
	logger           *logger.LoggerStruct
	rawAccountTotal  int
	uniqueIds        []string
	transactionQuery string
}

func MakeDatabase() *databaseStruct {
	dbInstance := databaseStruct{
		dbName:  "local-megalodon.db",
		dbUrl:   os.Getenv("TURSO_DATABASE_URL"),
		dbToken: os.Getenv("TURSO_AUTH_TOKEN"),
		logger:  logger.MakeLogger(),
	}

	dir, err := os.MkdirTemp("", "libsql-*")
	if err != nil {
		panic(err)
	}

	dbInstance.dbDir = dir
	dbInstance.dbPath = filepath.Join(dir, dbInstance.dbName)
	dbInstance.connect()

	return &dbInstance
}

func (db *databaseStruct) connect() {
	connector, err := libsql.NewEmbeddedReplicaConnector(db.dbPath, db.dbUrl, libsql.WithAuthToken(db.dbToken))
	if err != nil {
		panic(err)
	}

	db.dbConnector = connector
	db.client = sql.OpenDB(db.dbConnector)
}

func (db *databaseStruct) SyncAndClose() {
	db.logger.Info("Syncing database...")
	if _, err := db.dbConnector.Sync(); err != nil {
		panic(err)
	}

	db.logger.Info("Closing client...")
	if err := db.client.Close(); err != nil {
		db.logger.Error(err.Error())
	}

	db.logger.Info("Closing connector...")
	if err := db.dbConnector.Close(); err != nil {
		db.logger.Error(err.Error())
	}

	db.logger.Info("Cleaning temporary files...")
	os.RemoveAll(db.dbDir)
}

func (db *databaseStruct) createTableSafe() {
	var (
		crateTableQuery = `CREATE TABLE IF NOT EXISTS proxies (
			ID INTEGER PRIMARY KEY,
			SERVER TEXT,
			IP TEXT,
			SERVER_PORT INT8,
			UUID TEXT,
			PASSWORD TEXT,
			SECURITY TEXT,
			ALTER_ID INT2,
			METHOD TEXT,
			PLUGIN TEXT,
			PLUGIN_OPTS TEXT,
			HOST TEXT,
			TLS INT2,
			TRANSPORT TEXT,
			PATH TEXT,
			SERVICE_NAME TEXT,
			INSECURE INT2,
			SNI TEXT,
			REMARK TEXT,
			CONN_MODE TEXT,
			COUNTRY_CODE TEXT,
			REGION TEXT,
			ORG TEXT,
			VPN TEXT
		);`
	)

	if _, err := db.client.Query(crateTableQuery); err == nil {
		db.logger.Info("Database successfully created!")
	} else {
		db.logger.Error(err.Error())
		os.Exit(1)
	}
}

func (db *databaseStruct) Save(results []sandbox.TestResultStruct) error {
	db.createTableSafe()
	db.transactionQuery = db.buildInsertQuery(results)

	if _, err := db.client.Query(db.transactionQuery); err != nil {
		db.logger.Error(err.Error())
		bot.SendTextFileToAdmin(fmt.Sprintf("%v.txt", time.Now().Unix()), db.transactionQuery, err.Error())
		return err
	} else {
		db.logger.Info("=========================")
		db.logger.Success("Data successfully saved!")
		db.logger.Info(fmt.Sprintf("Total raw account: %d", db.rawAccountTotal))
		db.logger.Info(fmt.Sprintf("Total account saved: %d", len(db.uniqueIds)))
	}

	return nil
}

func (db *databaseStruct) buildInsertQuery(results []sandbox.TestResultStruct) string {
	var insertQuery = "BEGIN; DELETE FROM proxies;"
	db.rawAccountTotal = len(results)

	tableFieldValues := []DatabaseFieldStruct{}
	for _, result := range results {
		var (
			fieldValues = DatabaseFieldStruct{}
			outbound    = result.Outbound
		)

		var (
			outboundMapping = map[string]any{}
			outboundAny, _  = outbound.RawOptions()
			outboundByte, _ = json.Marshal(outboundAny)
		)
		json.Unmarshal(outboundByte, &outboundMapping)

		// Geoip
		fieldValues.Ip = result.ConfigGeoip.IP
		fieldValues.CountryCode = result.ConfigGeoip.Country
		fieldValues.Region = helper.GetRegionFromCC(fieldValues.CountryCode)
		fieldValues.Org = result.ConfigGeoip.AsOrganization

		// Common
		fieldValues.VPN = outbound.Type
		fieldValues.Server = outboundMapping["server"].(string)
		fieldValues.ServerPort = int(outboundMapping["server_port"].(float64))

		// Here we go assertion hell
		if uuid, ok := outboundMapping["uuid"].(string); ok {
			fieldValues.UUID = uuid
		}
		if password, ok := outboundMapping["password"].(string); ok {
			fieldValues.Password = password
		}
		if security, ok := outboundMapping["security"].(string); ok {
			fieldValues.Security = security
		}
		if alterId, ok := outboundMapping["alter_id"].(int); ok {
			fieldValues.AlterId = alterId
		}
		if method, ok := outboundMapping["method"].(string); ok {
			fieldValues.Method = method
		}
		if plugin, ok := outboundMapping["plugin"].(string); ok {
			fieldValues.Plugin = plugin
		}
		if pluginOpts, ok := outboundMapping["plugin_opts"].(string); ok {
			fieldValues.PluginOpts = pluginOpts
		}

		// Transport
		if outboundMapping["transport"] != nil {
			transportMapping := outboundMapping["transport"].(map[string]any)
			if transportType, ok := transportMapping["type"].(string); ok {
				fieldValues.Transport = transportType
			}
			if serviceName, ok := transportMapping["service_name"].(string); ok {
				fieldValues.ServiceName = serviceName
			}
			if path, ok := transportMapping["path"].(string); ok {
				fieldValues.Path = path
			}
			if host, ok := transportMapping["host"].(string); ok {
				fieldValues.Host = host
			}
			if transportMapping["headers"] != nil {
				headersMapping := transportMapping["headers"].(map[string]any)
				if host, ok := headersMapping["Host"].(string); ok {
					fieldValues.Host = host
				}
			}
		}

		// TLS
		tlsStr := "NTLS"
		if outboundMapping["tls"] != nil {
			tlsMapping := outboundMapping["tls"].(map[string]any)
			if enabled, ok := tlsMapping["enabled"].(bool); ok {
				fieldValues.TLS = enabled
				if enabled {
					tlsStr = "TLS"
				}
			}
			if insecure, ok := tlsMapping["insecure"].(bool); ok {
				fieldValues.Insecure = insecure
			}
			if sni, ok := tlsMapping["server_name"].(string); ok {
				fieldValues.SNI = sni
			}
		}

		for _, connMode := range result.TestPassed {
			fieldValues.ConnMode = connMode
			fieldValues.Remark = strings.ToUpper(fmt.Sprintf("%d %s %s %s %s %s", len(tableFieldValues)+1, helper.CCToEmoji(fieldValues.CountryCode), fieldValues.Org, fieldValues.Transport, connMode, tlsStr))

			// Check if same account exists
			if !db.checkIsExists(fieldValues) {
				tableFieldValues = append(tableFieldValues, fieldValues)
			}
		}
	}

	// Build queries
	values := []string{}
	for _, fieldValue := range tableFieldValues {
		value := "("

		value += fmt.Sprintf("'%s', ", fieldValue.Server)
		value += fmt.Sprintf("'%s', ", fieldValue.Ip)
		value += fmt.Sprintf("%d, ", fieldValue.ServerPort)
		value += fmt.Sprintf("'%s', ", fieldValue.UUID)
		value += fmt.Sprintf("'%s', ", fieldValue.Password)
		value += fmt.Sprintf("'%s', ", fieldValue.Security)
		value += fmt.Sprintf("%d, ", fieldValue.AlterId)
		value += fmt.Sprintf("'%v', ", fieldValue.Method)
		value += fmt.Sprintf("'%v', ", fieldValue.Plugin)
		value += fmt.Sprintf("'%v', ", fieldValue.PluginOpts)
		value += fmt.Sprintf("'%v', ", fieldValue.Host)
		value += fmt.Sprintf("%t, ", fieldValue.TLS)
		value += fmt.Sprintf("'%s', ", fieldValue.Transport)
		value += fmt.Sprintf("'%s', ", fieldValue.Path)
		value += fmt.Sprintf("'%s', ", fieldValue.ServiceName)
		value += fmt.Sprintf("%t, ", fieldValue.Insecure)
		value += fmt.Sprintf("'%s', ", fieldValue.SNI)
		value += fmt.Sprintf("'%s', ", fieldValue.Remark)
		value += fmt.Sprintf("'%s', ", fieldValue.ConnMode)
		value += fmt.Sprintf("'%s', ", fieldValue.CountryCode)
		value += fmt.Sprintf("'%s', ", fieldValue.Region)
		value += fmt.Sprintf("'%s', ", fieldValue.Org)
		value += fmt.Sprintf("'%s'", fieldValue.VPN)

		value += ")"

		values = append(values, value)
	}

	insertQuery += fmt.Sprintf(`INSERT INTO proxies (
		SERVER,
		IP,
		SERVER_PORT,
		UUID, PASSWORD,
		SECURITY,
		ALTER_ID,
		METHOD,
		PLUGIN,
		PLUGIN_OPTS,
		HOST,
		TLS,
		TRANSPORT,
		PATH,
		SERVICE_NAME,
		INSECURE,
		SNI,
		REMARK,
		CONN_MODE,
		COUNTRY_CODE,
		REGION,
		ORG,
		VPN
	) VALUES %s; COMMIT;`, strings.ReplaceAll(strings.Join(values, ","), `"`, ""))

	return insertQuery
}

func (db *databaseStruct) makeUniqueId(field DatabaseFieldStruct) string {
	// Server Port, UUID, Password, Plugin Opts, Path, Transport, Conn Mode, Country, Org, VPN
	uid := fmt.Sprintf("%d_%s_%s_%s_%s_%s_%s_%s_%s_%s", field.ServerPort, field.UUID, field.Password, field.PluginOpts, field.Path, field.Transport, field.ConnMode, field.CountryCode, field.Org, field.VPN)

	return uid
}

func (db *databaseStruct) checkIsExists(field DatabaseFieldStruct) bool {
	uid := db.makeUniqueId(field)
	for _, existedUid := range db.uniqueIds {
		if existedUid == uid {
			return true
		}
	}

	db.uniqueIds = append(db.uniqueIds, uid)
	return false
}
