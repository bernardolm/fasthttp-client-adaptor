package adaptor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetStatusText(t *testing.T) {
	assert.Equal(t, "200 OK", getStatusText(200))
	assert.Equal(t, "201 Created", getStatusText(201))
	assert.Equal(t, "400 Bad Request", getStatusText(400))
	assert.Equal(t, "404 Not Found", getStatusText(404))
	assert.Equal(t, "500 Internal Server Error", getStatusText(500))
	assert.Equal(t, "503 Service Unavailable", getStatusText(503))
	assert.Equal(t, "504 Gateway Timeout", getStatusText(504))
}

func TestThisBytesContains(t *testing.T) {
	assert.True(t, thisBytesContains(
		[]byte("Integer vitae felis in augue gravida condimentum sed eu nunc"),
		"felis"))
	assert.False(t, thisBytesContains(
		[]byte("Integer vitae felis in augue gravida condimentum sed eu nunc"),
		"neque"))

	assert.True(t, thisBytesContains(
		[]byte("Nunc non tortor non eros bibendum ullamcorper ac vel nunc"),
		"bend"))
	assert.False(t, thisBytesContains(
		[]byte("Nunc non tortor non eros bibendum ullamcorper ac vel nunc"),
		"erat"))
}
