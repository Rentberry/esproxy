package main

type Configuration struct {
	Debug         bool
	ListenAddress string `default:":19200"`
	ESAddress     string `envconfig:"ELASTICSEARCH_ADDRESS" default:"http://127.0.0.1:9200"`
	FlushInterval int    `envconfig:"FLUSH_INTERVAL" default:"20"`
}
