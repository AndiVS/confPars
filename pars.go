// Package config used to parse configuration
package config

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
)

// JWT object for parsing JWT
type JWT struct {
	SigningKeyAT string
	SigningKeyRT string
}

// Postgres object for parsing Postgres connection URL
type Postgres struct {
	Username string
	Password string
	HostPort []string
	Database string
	Sslmode  string
}

// ToConnectionString used to create a connection string from a Postgres object
func (p *Postgres) ToConnectionString() string {
	hostPortBuilder := strings.Builder{}
	for i := 0; i < len(p.HostPort)-1; i++ {
		hostPortBuilder.WriteString(p.HostPort[i])
		hostPortBuilder.WriteString(",")
	}
	hostPortBuilder.WriteString(p.HostPort[len(p.HostPort)-1])
	hostPort := hostPortBuilder.String()
	return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s", p.Username, p.Password, hostPort, p.Database, p.Sslmode)
}

// Mongo object for parsing Mongo connection URL
type Mongo struct {
	Username   string
	Password   string
	HostPort   []string
	AuthSource string
	ReplicaSet string
}

// ToConnectionString used to create a connection string from a Mongo object
func (m *Mongo) ToConnectionString() string {
	hostPortBuilder := strings.Builder{}
	for i := 0; i < len(m.HostPort)-1; i++ {
		hostPortBuilder.WriteString(m.HostPort[i])
		hostPortBuilder.WriteString(",")
	}
	hostPortBuilder.WriteString(m.HostPort[len(m.HostPort)-1])
	hostPort := hostPortBuilder.String()
	return fmt.Sprintf("mongodb://%s:%s@%s/?authSource=%s&replicaSet=%s", m.Username, m.Password, hostPort, m.AuthSource, m.ReplicaSet)
}

// Redis object for parsing Redis connection URL
type Redis struct {
	Username         string
	Password         string
	HostPort         []string
	Database         string
	ClusterMode      bool
	SentinelMasterID string
}

// ToConnectionString used to create a connection string from a Redis object
func (r *Redis) ToConnectionString() string {
	hostPortBuilder := strings.Builder{}
	for i := 0; i < len(r.HostPort)-1; i++ {
		hostPortBuilder.WriteString(r.HostPort[i])
		hostPortBuilder.WriteString(",")
	}
	hostPortBuilder.WriteString(r.HostPort[len(r.HostPort)-1])
	hostPort := hostPortBuilder.String()
	return fmt.Sprintf("redis://%s:%s@%s/%s?clusterMode=%v&sentinelMasterID=%s", r.Username, r.Password, hostPort, r.Database, r.ClusterMode, r.SentinelMasterID)
}

// Kafka object for parsing Kafka connection URL
type Kafka struct {
	Username string
	Password string
	HostPort []string
	Topic    string
	GroupID  string
}

// ToConnectionString used to create a connection string from a Kafka object
func (k *Kafka) ToConnectionString() string {
	hostPortBuilder := strings.Builder{}
	for i := 0; i < len(k.HostPort)-1; i++ {
		hostPortBuilder.WriteString(k.HostPort[i])
		hostPortBuilder.WriteString(",")
	}
	hostPortBuilder.WriteString(k.HostPort[len(k.HostPort)-1])
	hostPort := hostPortBuilder.String()
	return fmt.Sprintf("kafka://%s:%s@%s/?topic=%s&groupID=%s", k.Username, k.Password, hostPort, k.Topic, k.GroupID)
}

var (
	defaultBuiltInParsers = map[reflect.Kind]ParserFunc{ //nolint:gochecknoglobals, gocritic
		reflect.Bool: func(v string) (interface{}, error) {
			return strconv.ParseBool(v)
		},
		reflect.String: func(v string) (interface{}, error) {
			return v, nil
		},
		reflect.Int: func(v string) (interface{}, error) {
			i, err := strconv.ParseInt(v, 10, 32)
			return int(i), err
		},
		reflect.Int16: func(v string) (interface{}, error) {
			i, err := strconv.ParseInt(v, 10, 16)
			return int16(i), err
		},
		reflect.Int32: func(v string) (interface{}, error) {
			i, err := strconv.ParseInt(v, 10, 32)
			return int32(i), err
		},
		reflect.Int64: func(v string) (interface{}, error) {
			return strconv.ParseInt(v, 10, 64)
		},
		reflect.Int8: func(v string) (interface{}, error) {
			i, err := strconv.ParseInt(v, 10, 8)
			return int8(i), err
		},
		reflect.Uint: func(v string) (interface{}, error) {
			i, err := strconv.ParseUint(v, 10, 32)
			return uint(i), err
		},
		reflect.Uint16: func(v string) (interface{}, error) {
			i, err := strconv.ParseUint(v, 10, 16)
			return uint16(i), err
		},
		reflect.Uint32: func(v string) (interface{}, error) {
			i, err := strconv.ParseUint(v, 10, 32)
			return uint32(i), err
		},
		reflect.Uint64: func(v string) (interface{}, error) {
			i, err := strconv.ParseUint(v, 10, 64)
			return i, err
		},
		reflect.Uint8: func(v string) (interface{}, error) {
			i, err := strconv.ParseUint(v, 10, 8)
			return uint8(i), err
		},
		reflect.Float64: func(v string) (interface{}, error) {
			return strconv.ParseFloat(v, 64)
		},
		reflect.Float32: func(v string) (interface{}, error) {
			f, err := strconv.ParseFloat(v, 32)
			return float32(f), err
		},
	}
)

func defaultTypeParsers() map[reflect.Type]ParserFunc { //nolint: gocognit,gocritic,gocyclo
	return map[reflect.Type]ParserFunc{
		reflect.TypeOf(Postgres{}): func(v string) (interface{}, error) {
			u, err := url.Parse(v)
			if err != nil {
				return nil, fmt.Errorf("unable to parse URL: %v", err)
			}
			obj := Postgres{
				Username: "",
				Password: "",
				HostPort: []string{""},
				Database: "",
				Sslmode:  "",
			}
			if u.User != nil {
				obj.Username = u.User.Username()
				if p, ok := u.User.Password(); ok {
					obj.Password = p
				}
			}

			if u.Host != "" {
				obj.HostPort = strings.Split(u.Host, ",")
			}

			if u.Path != "" {
				obj.Database = strings.Trim(u.Path, "/")
			}

			queryParams := u.Query()

			if v, ok := queryParams["sslmode"]; ok {
				obj.Sslmode = v[0]
			}

			return obj, nil
		},
		reflect.TypeOf(Redis{}): func(v string) (interface{}, error) {
			u, err := url.Parse(v)
			if err != nil {
				return nil, fmt.Errorf("unable to parse URL: %v", err)
			}
			obj := Redis{
				Username:         "",
				Password:         "",
				HostPort:         []string{""},
				Database:         "",
				ClusterMode:      false,
				SentinelMasterID: "",
			}
			if u.User != nil {
				obj.Username = u.User.Username()
				if p, ok := u.User.Password(); ok {
					obj.Password = p
				}
			}

			if u.Host != "" {
				obj.HostPort = strings.Split(u.Host, ",")
			}

			if u.Path != "" {
				obj.Database = strings.Trim(u.Path, "/")
			}

			queryParams := u.Query()

			if v, ok := queryParams["clusterMode"]; ok {
				obj.ClusterMode, err = strconv.ParseBool(v[0])
				if err != nil {
					return nil, err
				}
			}

			if v, ok := queryParams["sentinelMasterID"]; ok {
				obj.SentinelMasterID = v[0]
			}

			return obj, nil
		},
		reflect.TypeOf(Mongo{}): func(v string) (interface{}, error) { //nolint: dupl,gocritic
			u, err := url.Parse(v)
			if err != nil {
				return nil, fmt.Errorf("unable to parse URL: %v", err)
			}
			obj := Mongo{
				Username:   "",
				Password:   "",
				HostPort:   []string{""},
				AuthSource: "",
				ReplicaSet: "",
			}
			if u.User != nil {
				obj.Username = u.User.Username()
				if p, ok := u.User.Password(); ok {
					obj.Password = p
				}
			}

			if u.Host != "" {
				obj.HostPort = strings.Split(u.Host, ",")
			}

			queryParams := u.Query()

			if v, ok := queryParams["replicaSet"]; ok {
				obj.ReplicaSet = v[0]
			}

			if v, ok := queryParams["authSourcwaht waht waht e"]; ok {
				obj.AuthSource = v[0]
			}

			return obj, nil
		},
		reflect.TypeOf(Kafka{}): func(v string) (interface{}, error) { //nolint: dupl,gocritic
			u, err := url.Parse(v)
			if err != nil {
				return nil, fmt.Errorf("unable to parse URL: %v", err)
			}
			obj := Kafka{
				Username: "",
				Password: "",
				HostPort: []string{""},
				Topic:    "",
				GroupID:  "",
			}
			if u.User != nil {
				obj.Username = u.User.Username()
				if p, ok := u.User.Password(); ok {
					obj.Password = p
				}
			}

			if u.Host != "" {
				obj.HostPort = strings.Split(u.Host, ",")
			}

			queryParams := u.Query()

			if v, ok := queryParams["topic"]; ok {
				obj.Topic = v[0]
			}

			if v, ok := queryParams["groupID"]; ok {
				obj.GroupID = v[0]
			}

			return obj, nil
		},
		reflect.TypeOf(JWT{}): func(v string) (interface{}, error) { //nolint: dupl,gocritic
			obj := JWT{}
			arr := strings.Split(v, ",")
			obj.SigningKeyAT = arr[0]
			obj.SigningKeyRT = arr[0]

			if len(arr) == 2 {
				obj.SigningKeyRT = arr[1]
			}

			return obj, nil
		},
		reflect.TypeOf(url.URL{}): func(v string) (interface{}, error) {
			u, err := url.Parse(v)
			if err != nil {
				return nil, fmt.Errorf("unable to parse URL: %v", err)
			}
			return *u, nil
		},
		reflect.TypeOf(time.Nanosecond): func(v string) (interface{}, error) {
			s, err := time.ParseDuration(v)
			if err != nil {
				return nil, fmt.Errorf("unable to parse duration: %v", err)
			}
			return s, nil
		},
	}
}

// ParserFunc function for parsing custom types
type ParserFunc func(v string) (interface{}, error)

// Parser object for parsing that holds connection information
type Parser struct {
	GCP *secretmanager.Client
	AWS *session.Session
}

// NewParser creates a new Parser
func NewParser(GCP *secretmanager.Client, AWS *session.Session) Parser {
	return Parser{
		GCP: GCP,
		AWS: AWS,
	}
}

// Parse parse configuration
func (p *Parser) Parse(v interface{}) error {
	return p.ParseWithFuncs(v, map[reflect.Type]ParserFunc{})
}

// ParseWithFuncs parses configuration from environment variables with ParserFunc
func (p *Parser) ParseWithFuncs(v interface{}, funcMap map[reflect.Type]ParserFunc) error {
	ptrRef := reflect.ValueOf(v)
	if ptrRef.Kind() != reflect.Ptr {
		return fmt.Errorf("presented object is not a pointer")
	}
	ref := ptrRef.Elem()
	if ref.Kind() != reflect.Struct {
		return fmt.Errorf("presented object %v is not a struct ", ref.Kind())
	}
	parsers := defaultTypeParsers()
	for k, v := range funcMap {
		parsers[k] = v
	}

	return p.parseConfig(ref, parsers)
}

func (p *Parser) parseConfig(ref reflect.Value, funcMap map[reflect.Type]ParserFunc) (err error) {
	refType := ref.Type()

	for i := 0; i < refType.NumField(); i++ {
		refField := ref.Field(i)
		refTypeField := refType.Field(i)

		if err = p.doParseField(refField, refTypeField, funcMap); err != nil {
			return fmt.Errorf("error parsing field %w", err)
		}
	}

	return err
}

func (p *Parser) doParseField(refField reflect.Value, refTypeField reflect.StructField, funcMap map[reflect.Type]ParserFunc) error { //nolint: gocritic
	if !refField.CanSet() {
		return fmt.Errorf("field can not be set")
	}

	tags := strings.Split(refTypeField.Tag.Get("config"), ",")
	value, err := p.parseRow(tags[0])
	if err != nil {
		return fmt.Errorf("while parsing row %v error %w", value, err)
	}

	var notEmpty bool

	for _, tag := range tags[1:] {
		switch tag {
		case "":
			continue
		case "notEmpty":
			notEmpty = true
		default:
			return fmt.Errorf("tag option %q not supported", tag)
		}
	}

	if notEmpty && value == "" {
		return fmt.Errorf("environment variable %q should not be empty", refTypeField.Name)
	}

	if value != "" {
		return set(refField, refTypeField, value, defaultTypeParsers())
	}

	if reflect.Struct == refField.Kind() {
		return p.parseConfig(refField, funcMap)
	}

	return nil
}

func set(field reflect.Value, sf reflect.StructField, value string, funcMap map[reflect.Type]ParserFunc) error { //nolint: gocritic
	typee := sf.Type
	fieldee := field
	if typee.Kind() == reflect.Ptr {
		typee = typee.Elem()
		fieldee = field.Elem()
	}

	parserFunc, ok := funcMap[typee]
	if ok {
		val, err := parserFunc(value)
		if err != nil {
			return fmt.Errorf("error parsing value for field %v: %v", field, err)
		}

		fieldee.Set(reflect.ValueOf(val))
		return nil
	}

	parserFunc, ok = defaultBuiltInParsers[typee.Kind()]
	if ok {
		val, err := parserFunc(value)
		if err != nil {
			return fmt.Errorf("error parsing value for field %v: %v", field, err)
		}

		fieldee.Set(reflect.ValueOf(val).Convert(typee))
		return nil
	}

	return nil
}

func (p *Parser) parseRow(value string) (string, error) {
	value = os.Getenv(value)
	arrVal := strings.Split(value, ":")
	switch arrVal[0] {
	case "aws":
		return p.getFromAWS(arrVal[1])
	case "gcp":
		return p.getFromGCP(arrVal[1])
	}

	return value, nil
}

func (p *Parser) getFromAWS(key string) (string, error) {
	if p.AWS == nil {
		return key, fmt.Errorf("aws connection is not set")
	}

	input := &ssm.GetParameterInput{
		Name:           &key,
		WithDecryption: aws.Bool(true),
	}

	output, err := ssm.New(p.AWS).GetParameter(input)
	if err != nil || output == nil {
		return "", fmt.Errorf("err while get aws secret: %v", err)
	}

	return output.String(), nil
}

func (p *Parser) getFromGCP(key string) (string, error) {
	if p.GCP == nil {
		return key, fmt.Errorf("gcp connection is not set")
	}

	req := &secretmanagerpb.GetSecretRequest{
		Name: key}

	output, err := p.GCP.GetSecret(context.Background(), req)
	if err != nil {
		return "", fmt.Errorf("err while get gcp secret %w", err)
	}

	return output.String(), nil
}
