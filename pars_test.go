package config

import (
	"os"
	"testing"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/stretchr/testify/require"
)

func TestParser_Parse(t *testing.T) {
	type fields struct {
		GCP *secretmanager.Client
		AWS *session.Session
	}
	tests := []struct {
		name      string
		fields    fields
		args      interface{}
		wantValue interface{}
		wantErr   bool
		testData  string
		testKey   string
	}{
		{
			name: "default_string_not_empty_error",
			fields: fields{
				GCP: nil,
				AWS: nil,
			},
			args: &struct {
				TestStr string `config:"string,notEmpty"`
			}{},
			wantErr: true,
			wantValue: &struct {
				TestStr string `config:"string,notEmpty"`
			}{TestStr: "1111"},
			testData: "",
			testKey:  "string",
		},
		{
			name: "default_string_not_empty",
			fields: fields{
				GCP: nil,
				AWS: nil,
			},
			args: &struct {
				TestStr string `config:"string,notEmpty"`
			}{},
			wantErr: false,
			wantValue: &struct {
				TestStr string `config:"string,notEmpty"`
			}{TestStr: "1111"},
			testData: "1111",
			testKey:  "string",
		},
		{
			name: "kafka",
			fields: fields{
				GCP: nil,
				AWS: nil,
			},
			args: &struct {
				Kafka `config:"kafka"`
			}{},
			wantErr: false,
			wantValue: &struct {
				Kafka `config:"kafka"`
			}{Kafka{
				Username: "username",
				Password: "password",
				HostPort: []string{"localhost1:1111", "localhost2:2222"},
				Topic:    "topic",
				GroupID:  "groupID",
			}},
			testData: "kafka://username:password@localhost1:1111,localhost2:2222/?topic=topic&groupID=groupID",
			testKey:  "kafka",
		},
		{
			name: "redis",
			fields: fields{
				GCP: nil,
				AWS: nil,
			},
			args: &struct {
				Redis `config:"redis"`
			}{},
			wantErr: false,
			wantValue: &struct {
				Redis `config:"redis"`
			}{Redis{
				Username:         "username",
				Password:         "password",
				HostPort:         []string{"localhost1:1111", "localhost2:2222"},
				Database:         "database",
				ClusterMode:      false,
				SentinelMasterID: "sentinelMasterID",
			}},
			testData: "redis://username:password@localhost1:1111,localhost2:2222/database?clusterMode=false&sentinelMasterID=sentinelMasterID",
			testKey:  "redis",
		},
		{
			name: "postgres",
			fields: fields{
				GCP: nil,
				AWS: nil,
			},
			args: &struct {
				Postgres `config:"postgres"`
			}{},
			wantErr: false,
			wantValue: &struct {
				Postgres `config:"postgres"`
			}{Postgres{
				Username: "username",
				Password: "password",
				HostPort: []string{"localhost1:1111", "localhost2:2222"},
				Database: "database",
				Sslmode:  "disable",
			}},
			testData: "postgres://username:password@localhost1:1111,localhost2:2222/database?sslmode=disable",
			testKey:  "postgres",
		},
		{
			name: "мongo",
			fields: fields{
				GCP: nil,
				AWS: nil,
			},
			args: &struct {
				Mongo `config:"мongo"`
			}{},
			wantErr: false,
			wantValue: &struct {
				Mongo `config:"мongo"`
			}{Mongo{
				Username:   "username",
				Password:   "password",
				HostPort:   []string{"localhost1:1111", "localhost2:2222"},
				AuthSource: "admin",
				ReplicaSet: "myRepl",
			}},
			testData: "mongodb://username:password@localhost1:1111,localhost2:2222/?authSource=admin&replicaSet=myRepl",
			testKey:  "мongo",
		},
		{
			name: "JWT",
			fields: fields{
				GCP: nil,
				AWS: nil,
			},
			args: &struct {
				JWT `config:"jwt"`
			}{},
			wantErr: false,
			wantValue: &struct {
				JWT `config:"jwt"`
			}{JWT{
				SigningKeyAT: "SigningKeyAT",
				SigningKeyRT: "SigningKeyRT",
			}},
			testData: "SigningKeyAT,SigningKeyRT",
			testKey:  "jwt",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := os.Setenv(tt.testKey, tt.testData)
			require.NoError(t, err)
			p := &Parser{
				GCP: tt.fields.GCP,
				AWS: tt.fields.AWS,
			}
			gotErr := p.Parse(tt.args)
			if tt.wantErr {
				require.Error(t, gotErr, "")
			} else {
				require.NoError(t, gotErr)
				require.EqualValues(t, tt.wantValue, tt.args)
			}
		})
	}
}

func TestPostgres_ToConnectionString(t *testing.T) {
	type fields struct {
		Username string
		Password string
		HostPort []string
		Database string
		Sslmode  string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "pass",
			fields: fields{
				Username: "username",
				Password: "password",
				HostPort: []string{"localhost1:1111", "localhost2:2222"},
				Database: "database",
				Sslmode:  "disable",
			},
			want: "postgres://username:password@localhost1:1111,localhost2:2222/database?sslmode=disable",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Postgres{
				Username: tt.fields.Username,
				Password: tt.fields.Password,
				HostPort: tt.fields.HostPort,
				Database: tt.fields.Database,
				Sslmode:  tt.fields.Sslmode,
			}
			got := p.ToConnectionString()
			require.Equal(t, tt.want, got, "ToConnectionString() = %v, want %v", got, tt.want)
		})
	}
}

func TestMongo_ToConnectionString(t *testing.T) {
	type fields struct {
		Username   string
		Password   string
		HostPort   []string
		AuthSource string
		ReplicaSet string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "pass",
			fields: fields{
				Username:   "username",
				Password:   "password",
				HostPort:   []string{"localhost1:1111", "localhost2:2222"},
				AuthSource: "admin",
				ReplicaSet: "myRepl",
			},
			want: "mongodb://username:password@localhost1:1111,localhost2:2222/?authSource=admin&replicaSet=myRepl",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Mongo{
				Username:   tt.fields.Username,
				Password:   tt.fields.Password,
				HostPort:   tt.fields.HostPort,
				AuthSource: tt.fields.AuthSource,
				ReplicaSet: tt.fields.ReplicaSet,
			}
			got := m.ToConnectionString()
			require.Equal(t, tt.want, got, "ToConnectionString() = %v, want %v", got, tt.want)
		})
	}
}

func TestRedis_ToConnectionString(t *testing.T) {
	type fields struct {
		Username         string
		Password         string
		HostPort         []string
		Database         string
		ClusterMode      bool
		SentinelMasterID string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "pass",
			fields: fields{
				Username:         "username",
				Password:         "password",
				HostPort:         []string{"localhost1:1111", "localhost2:2222"},
				Database:         "database",
				ClusterMode:      false,
				SentinelMasterID: "sentinelMasterID",
			},
			want: "redis://username:password@localhost1:1111,localhost2:2222/database?clusterMode=false&sentinelMasterID=sentinelMasterID",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Redis{
				Username:         tt.fields.Username,
				Password:         tt.fields.Password,
				HostPort:         tt.fields.HostPort,
				Database:         tt.fields.Database,
				ClusterMode:      tt.fields.ClusterMode,
				SentinelMasterID: tt.fields.SentinelMasterID,
			}
			got := r.ToConnectionString()
			require.Equal(t, tt.want, got, "ToConnectionString() = %v, want %v", got, tt.want)
		})
	}
}

func TestKafka_ToConnectionString(t *testing.T) {
	type fields struct {
		Username string
		Password string
		HostPort []string
		Topic    string
		GroupID  string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "pass",
			fields: fields{
				Username: "username",
				Password: "password",
				HostPort: []string{"localhost1:1111", "localhost2:2222"},
				Topic:    "topic",
				GroupID:  "groupID",
			},
			want: "kafka://username:password@localhost1:1111,localhost2:2222/?topic=topic&groupID=groupID",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &Kafka{
				Username: tt.fields.Username,
				Password: tt.fields.Password,
				HostPort: tt.fields.HostPort,
				Topic:    tt.fields.Topic,
				GroupID:  tt.fields.GroupID,
			}
			got := k.ToConnectionString()
			require.Equal(t, tt.want, got, "ToConnectionString() = %v, want %v", got, tt.want)
		})
	}
}
