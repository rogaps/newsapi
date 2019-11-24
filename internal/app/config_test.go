package app

import "testing"

import "github.com/stretchr/testify/assert"

func TestParse(t *testing.T) {
	config := &Config{}
	config.Parse("../../configs/config-newsservice.json")
	assert.Equal(t, config.Server.Address, ":8080")
	assert.Equal(t, config.Elasticsearch.ConnectionString, []string{"http://localhost:9200"})
	assert.Equal(t, config.AMQP.ConnectionString, "amqp://johndoe:pwd0123456789@localhost:5672")
}
