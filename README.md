# config

This package can be used to parse the configuration from environment variables, the GCP secret manager and AWS SSM.
Use tag "config" to specify key for environment variable that will be parsed
also to parse configuration form aws secret manager set aws session using newParser function
like 
 {
    PostgresConnectionString string 'config:"POSTGRES_CONNECTION_STRING_KEY"'
            where POSTGRES_CONNECTION_STRING_KEY = "postgres://username:password@localhost1:1111,localhost2:2222/database?sslmode=disable"
                                                or "gcp:"postgres key form gcp key storage""
}

For JWT, Redis, Kafka, Postgres and Mongo can be used structure form current package

examples:

    any other default type : "string"
    
    jwt:            "SigningKeyAT,SigningKeyRT"
    redis url:      "redis://username:password@localhost1:1111,localhost2:2222/database?clusterMode=false&sentinelMasterID=sentinelMasterID"
    kafka url:      "kafka://username:password@localhost1:1111,localhost2:2222/?topic=topic&groupID=groupID"
    postgres url:   "postgres://username:password@localhost1:1111,localhost2:2222/database?sslmode=disable"
    mongo url:      "mongodb://username:password@localhost1:1111,localhost2:2222/?authSource=admin&replicaSet=myRepl"

To get the configuration from GCP, set the `config` tag value as "gcp:secret key you want to get".

    any other default type : "gcp:gcp_key"  

    jwt:            "gcp:gcp_key"  
    redis url:      "gcp:gcp_key"  
    kafka url:      "gcp:gcp_key"
    postgres url:   "gcp:gcp_key"
    mongo url:      "gcp:gcp_key"

on GCP under the key, connection string provided, as shown in the first example.

To retrieve the configuration from AWS, set the `config` tag to "aws:secret key you want to retrieve".

    any other default type : "aws:aws_key"

    jwt:            "aws:aws_key" 
    redis url:      "aws:aws_key"  
    kafka url:      "aws:aws_key"
    postgres url:   "aws:aws_key"
    mongo url:      "aws:aws_key"

on AWS under the presented key, connection string, as shown in the first example