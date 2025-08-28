package command_test

import (
	"testing"
	"trace-monitor-collector/command"

	"github.com/stretchr/testify/assert"
)

func TestFromJsonReturnsErrorWhenJsonIsEmpty(t *testing.T) {
	// Arrange
	js := []byte(``)

	// Act
	cmd, err := command.FromJson(js)

	// Assert
	assert.ErrorIs(t, err, command.ErrorEmptyJson)
	assert.Empty(t, cmd)
}

func TestFromJsonReturnsErrorWhenJsonIsInvalid(t *testing.T) {
	// Arrange
	js := []byte(`{"invalid json"}`)

	// Act
	cmd, err := command.FromJson(js)

	// Assert
	assert.NotNil(t, err)
	assert.Empty(t, cmd)
}

func TestFromJsonReturnsValidCommandWhenJsonIsValidAndOfRightStructure(t *testing.T) {
	// Arrange
	js := []byte(`{"method":"set-current-span","sentAt":"2023-04-10T14:04:31.387230+03:00","pid":"2905890","traceId":"0bbf9e15-519d-4e4f-af14-eb4caa40e88b","data":{"span":{"id":"b17482a2-5694-434b-9133-8c3c6ee81a0d","parent":null,"openedAt":"2023-04-10T14:04:31.387220+03:00","name":"Database query","context":{"query":"select * from \"available_for_rent_cars\" where \"available_for_rent_cars\".\"id\" = ? limit 1","bindings":[9123]},"tags":[],"debugTrace":[{"file":"\/home\/www-data\/backend-rent\/releases\/130856_670596\/vendor\/laravel\/framework\/src\/Illuminate\/Database\/Connection.php","line":624,"function":"runQueryCallback","class":"BelkaCar\\LaravelIntegration\\TraceMonitor\\Database\\TraceMonitorPostgisConnection","type":"->"},{"file":"\/home\/www-data\/backend-rent\/releases\/130856_670596\/vendor\/laravel\/framework\/src\/Illuminate\/Database\/Connection.php","line":333,"function":"run","class":"Illuminate\\Database\\Connection","type":"->"},{"file":"\/home\/www-data\/backend-rent\/releases\/130856_670596\/vendor\/laravel\/framework\/src\/Illuminate\/Database\/Query\/Builder.php","line":1719,"function":"select","class":"Illuminate\\Database\\Connection","type":"->"}]},"parentSpans":[]}}`)

	// Act
	cmd, err := command.FromJson(js)

	// Assert
	assert.Nil(t, err)
	assert.Equal(t, js, cmd.RawCommand)
	assert.Equal(t, "2905890", cmd.Pid)
	assert.Equal(t, "set-current-span", cmd.Method)
	assert.Equal(t, "0bbf9e15-519d-4e4f-af14-eb4caa40e88b", cmd.TraceId)
	assert.Equal(t, "2023-04-10T14:04:31.387230+03:00", cmd.SentAt.Format("2006-01-02T15:04:05.000000-07:00"))

}
