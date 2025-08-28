package main

import (
	"context"
	"net"
	"testing"
	"time"
	"trace-monitor-collector/config"
	"trace-monitor-collector/traceCollection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatsCalculatedCorrectly(t *testing.T) {
	// Arrange
	ctx, cancel := context.WithCancel(context.Background())
	go handleUdp(ctx, udpServerConfig(8080))
	<-udpServerReadyChan

	udpServer, err := net.ResolveUDPAddr("udp", ":8080")
	require.Nil(t, err)
	client, err := net.DialUDP("udp", nil, udpServer)
	require.Nil(t, err)
	defer client.Close()

	packets := []string{
		`{"method":"init-trace","sentAt":"2023-04-10T14:04:31.367337+03:00","pid":"2905890","traceId":"0bbf9e15-519d-4e4f-af14-eb4caa40e88b","data":{"serverContext":{"ip":"172.16.101.40","method":"GET","host":"mapi.belkacar.ru","uri":"\/v3.0\/car-address?car_id=9123","body_input":"","body_post":"[]"},"context":[],"tags":[],"openedAt":"2023-04-10T14:04:31.367318+03:00"}}`,
		`{"method":"set-trace-current-span","sentAt":"2023-04-10T14:04:31.375059+03:00","pid":"2905890","traceId":"0bbf9e15-519d-4e4f-af14-eb4caa40e88b","data":{"span":{"id":"211cadad-4c63-46b7-a2ac-fe68f735f4f0","parent":null,"openedAt":"2023-04-10T14:04:31.375036+03:00","name":"Database query","context":{"query":"select * from \"user_device\" where \"device_id\" = ? and (\"auth_token\" = ? or \"previous_auth_token\" = ?) and (\"user_device\".\"deleted_at\") is null limit 1","bindings":["A7526423-BD21-4DA1-9954-C172ACEB96DB","4b6faee95eae2ab68fb472b2d2de627fa0b621f2","4b6faee95eae2ab68fb472b2d2de627fa0b621f2"]},"tags":[],"debugTrace":[{"file":"\/home\/www-data\/backend-rent\/releases\/130856_670596\/vendor\/laravel\/framework\/src\/Illuminate\/Database\/Connection.php","line":624,"function":"runQueryCallback","class":"BelkaCar\\LaravelIntegration\\TraceMonitor\\Database\\TraceMonitorPostgisConnection","type":"->"},{"file":"\/home\/www-data\/backend-rent\/releases\/130856_670596\/vendor\/laravel\/framework\/src\/Illuminate\/Database\/Connection.php","line":333,"function":"run","class":"Illuminate\\Database\\Connection","type":"->"},{"file":"\/home\/www-data\/backend-rent\/releases\/130856_670596\/vendor\/laravel\/framework\/src\/Illuminate\/Database\/Query\/Builder.php","line":1719,"function":"select","class":"Illuminate\\Database\\Connection","type":"->"}]},"parentSpans":[]}}`,
		`{"method":"set-trace-current-span","sentAt":"2023-04-10T14:04:31.376382+03:00","pid":"2905890","traceId":"0bbf9e15-519d-4e4f-af14-eb4caa40e88b","data":null}`,
		`{"method":"set-trace-current-span","sentAt":"2023-04-10T14:04:31.379728+03:00","pid":"2905890","traceId":"0bbf9e15-519d-4e4f-af14-eb4caa40e88b","data":{"span":{"id":"34af079c-0668-45f4-a09b-e6c35a0b4e7a","parent":null,"openedAt":"2023-04-10T14:04:31.379720+03:00","name":"Database query","context":{"query":"select * from \"user\" where (\"deleted_at\") is null and \"user_role\" = ? and \"user\".\"id\" = ? limit 1","bindings":["basic",3824616]},"tags":[],"debugTrace":[{"file":"\/home\/www-data\/backend-rent\/releases\/130856_670596\/vendor\/laravel\/framework\/src\/Illuminate\/Database\/Connection.php","line":624,"function":"runQueryCallback","class":"BelkaCar\\LaravelIntegration\\TraceMonitor\\Database\\TraceMonitorPostgisConnection","type":"->"},{"file":"\/home\/www-data\/backend-rent\/releases\/130856_670596\/vendor\/laravel\/framework\/src\/Illuminate\/Database\/Connection.php","line":333,"function":"run","class":"Illuminate\\Database\\Connection","type":"->"},{"file":"\/home\/www-data\/backend-rent\/releases\/130856_670596\/vendor\/laravel\/framework\/src\/Illuminate\/Database\/Query\/Builder.php","line":1719,"function":"select","class":"Illuminate\\Database\\Connection","type":"->"}]},"parentSpans":[]}}`,
		`{"method":"set-trace-current-span","sentAt":"2023-04-10T14:04:31.381308+03:00","pid":"2905890","traceId":"0bbf9e15-519d-4e4f-af14-eb4caa40e88b","data":null}`,
		`{"method":"set-trace-current-span","sentAt":"2023-04-10T14:04:31.382452+03:00","pid":"2905890","traceId":"0bbf9e15-519d-4e4f-af14-eb4caa40e88b","data":{"span":{"id":"69c5d2e7-54d7-44d7-bd6e-843a81be4795","parent":null,"openedAt":"2023-04-10T14:04:31.382445+03:00","name":"Redis command","context":{"commandID":"eval","arguments":["local userIpsKey = KEYS[1]\nlocal userIpChangeBlockKey = KEYS[2]\nlocal userEndpointRequestKey = KEYS[3]\nlocal userEndpointRequestBlockKey = KEYS[4]\n\nlocal ip = ARGV[1]\nlocal userIpCountLimit = tonumber(ARGV[2])\nlocal userIpsKeyTTL = tonumber(ARGV[3])\nlocal userIpChangeBlockKeyTTL = tonumber(ARGV[4])\nlocal endpointMask = ARGV[5]\nlocal useCurrentEndpointLimit = tonumber(ARGV[6])\nlocal userEndpointRequestKeyTTL = tonumber(ARGV[7])\nlocal userEndpointRequestBlockKeyTTL = tonumber(ARGV[8])\n\n-- save reference information\nredis.call('hincrby', userIpsKey, ip, 1)\nlocal currentUserIpsKeyTTL = redis.call('ttl', userIpsKey)\nif currentUserIpsKeyTTL == -1 then\n    redis.call('expire', userIpsKey, userIpsKeyTTL)\nend\n\nredis.call('hincrby', userEndpointRequestKey, endpointMask, 1)\nlocal currentUserEndpointsKeyTTL = redis.call('ttl', userEndpointRequestKey)\nif currentUserEndpointsKeyTTL == -1 then\n    redis.call('expire', userEndpointRequestKey, userEndpointRequestKeyTTL)\nend\n--\n\n-- check user blocking f...",4,"user_throttler:user_ips:3824616","user_throttler:ip_change_blocked_user:3824616","user_throttler:user_endpoints:3824616","user_throttler:endpoint_request_blocked_user:3824616","213.87.89.109",6,240,3600,"get:car-address",-1,86400,60],"config":{"host":"rent-redis-01.belkacar.ru","password":null,"port":6379,"database":"3"},"formattedOptions":{"timeout":10}},"tags":[],"debugTrace":[{"file":"\/home\/www-data\/backend-rent\/releases\/130856_670596\/belkacar\/Mapi\/src\/Throttler\/UserThrottlerService.php","line":161,"function":"__call","class":"class@anonymous\u0000\/home\/www-data\/backend-rent\/releases\/130856_670596\/vendor\/belkacar\/laravel-integration\/src\/IntegrationServiceProvider.php:120$48","type":"->"},{"file":"\/home\/www-data\/backend-rent\/releases\/130856_670596\/belkacar\/Mapi\/src\/Throttler\/UserThrottlerService.php","line":100,"function":"eval","class":"BelkaCar\\Mapi\\Throttler\\UserThrottlerService","type":"->"},{"file":"\/home\/www-data\/backend-rent\/releases\/130856_670596\/belkacar\/Mapi\/src\/Service\/ThrottlerService.php","line":84,"function":"throttle","class":"BelkaCar\\Mapi\\Throttler\\UserThrottlerService","type":"->"}]},"parentSpans":[]}}`,
		`{"method":"set-trace-current-span","sentAt":"2023-04-10T14:04:31.383338+03:00","pid":"2905890","traceId":"0bbf9e15-519d-4e4f-af14-eb4caa40e88b","data":null}`,
		`{"method":"set-trace-current-span","sentAt":"2023-04-10T14:04:31.383879+03:00","pid":"2905890","traceId":"0bbf9e15-519d-4e4f-af14-eb4caa40e88b","data":{"span":{"id":"bbe6f246-4727-49a8-9354-43c8ff461f52","parent":null,"openedAt":"2023-04-10T14:04:31.383872+03:00","name":"Redis command","context":{"commandID":"exists","arguments":["user_session:A7526423-BD21-4DA1-9954-C172ACEB96DB"],"config":{"host":"rent-redis-01.belkacar.ru","password":null,"port":6379,"database":"2"},"formattedOptions":{"timeout":10}},"tags":[],"debugTrace":[{"file":"\/home\/www-data\/backend-rent\/releases\/130856_670596\/belkacar\/Mapi\/src\/Service\/MapiSessionService.php","line":59,"function":"__call","class":"class@anonymous\u0000\/home\/www-data\/backend-rent\/releases\/130856_670596\/vendor\/belkacar\/laravel-integration\/src\/IntegrationServiceProvider.php:120$48","type":"->"},{"file":"\/home\/www-data\/backend-rent\/releases\/130856_670596\/belkacar\/Mapi\/src\/Middleware\/Mapi30Middleware.php","line":123,"function":"touchUserSession","class":"BelkaCar\\Mapi\\Service\\MapiSessionService","type":"->"},{"file":"\/home\/www-data\/backend-rent\/releases\/130856_670596\/belkacar\/Mapi\/src\/Middleware\/Mapi30Middleware.php","line":74,"function":"touchUserSession","class":"BelkaCar\\Mapi\\Middleware\\Mapi30Middleware","type":"->"}]},"parentSpans":[]}}`,
		`{"method":"set-trace-current-span","sentAt":"2023-04-10T14:04:31.384833+03:00","pid":"2905890","traceId":"0bbf9e15-519d-4e4f-af14-eb4caa40e88b","data":null}`,
		`{"method":"set-trace-current-span","sentAt":"2023-04-10T14:04:31.385133+03:00","pid":"2905890","traceId":"0bbf9e15-519d-4e4f-af14-eb4caa40e88b","data":{"span":{"id":"b83c56e7-41bf-45e6-85a6-8652b8dcd2d3","parent":null,"openedAt":"2023-04-10T14:04:31.385125+03:00","name":"Redis command","context":{"commandID":"setex","arguments":["user_session:A7526423-BD21-4DA1-9954-C172ACEB96DB",180,"{\"mobile_type\":\"new_app\",\"mobile_build\":\"ios 2.10.7 1001\",\"mobile_platform\":\"ios\",\"is_mapi\":\"true\"}"],"config":{"host":"rent-redis-01.belkacar.ru","password":null,"port":6379,"database":"2"},"formattedOptions":{"timeout":10}},"tags":[],"debugTrace":[{"file":"\/home\/www-data\/backend-rent\/releases\/130856_670596\/belkacar\/Mapi\/src\/Service\/MapiSessionService.php","line":63,"function":"__call","class":"class@anonymous\u0000\/home\/www-data\/backend-rent\/releases\/130856_670596\/vendor\/belkacar\/laravel-integration\/src\/IntegrationServiceProvider.php:120$48","type":"->"},{"file":"\/home\/www-data\/backend-rent\/releases\/130856_670596\/belkacar\/Mapi\/src\/Middleware\/Mapi30Middleware.php","line":123,"function":"touchUserSession","class":"BelkaCar\\Mapi\\Service\\MapiSessionService","type":"->"},{"file":"\/home\/www-data\/backend-rent\/releases\/130856_670596\/belkacar\/Mapi\/src\/Middleware\/Mapi30Middleware.php","line":74,"function":"touchUserSession","class":"BelkaCar\\Mapi\\Middleware\\Mapi30Middleware","type":"->"}]},"parentSpans":[]}}`,
		`{"method":"set-trace-current-span","sentAt":"2023-04-10T14:04:31.385640+03:00","pid":"2905890","traceId":"0bbf9e15-519d-4e4f-af14-eb4caa40e88b","data":null}`,
		`{"method":"set-trace-current-span","sentAt":"2023-04-10T14:04:31.387230+03:00","pid":"2905890","traceId":"0bbf9e15-519d-4e4f-af14-eb4caa40e88b","data":{"span":{"id":"b17482a2-5694-434b-9133-8c3c6ee81a0d","parent":null,"openedAt":"2023-04-10T14:04:31.387220+03:00","name":"Database query","context":{"query":"select * from \"available_for_rent_cars\" where \"available_for_rent_cars\".\"id\" = ? limit 1","bindings":[9123]},"tags":[],"debugTrace":[{"file":"\/home\/www-data\/backend-rent\/releases\/130856_670596\/vendor\/laravel\/framework\/src\/Illuminate\/Database\/Connection.php","line":624,"function":"runQueryCallback","class":"BelkaCar\\LaravelIntegration\\TraceMonitor\\Database\\TraceMonitorPostgisConnection","type":"->"},{"file":"\/home\/www-data\/backend-rent\/releases\/130856_670596\/vendor\/laravel\/framework\/src\/Illuminate\/Database\/Connection.php","line":333,"function":"run","class":"Illuminate\\Database\\Connection","type":"->"},{"file":"\/home\/www-data\/backend-rent\/releases\/130856_670596\/vendor\/laravel\/framework\/src\/Illuminate\/Database\/Query\/Builder.php","line":1719,"function":"select","class":"Illuminate\\Database\\Connection","type":"->"}]},"parentSpans":[]}}`,
		`{"method":"set-trace-current-span","sentAt":"2023-04-10T14:04:31.388668+03:00","pid":"2905890","traceId":"0bbf9e15-519d-4e4f-af14-eb4caa40e88b","data":null}`,
		`{"method":"set-trace-current-span","sentAt":"2023-04-10T14:04:31.389430+03:00","pid":"2905890","traceId":"0bbf9e15-519d-4e4f-af14-eb4caa40e88b","data":{"span":{"id":"c7652386-4555-4d57-a7c6-8089a2d101bd","parent":null,"openedAt":"2023-04-10T14:04:31.389417+03:00","name":"Database query","context":{"query":"select * from \"rent\" where \"car_id\" = ? and (\"finished_at\") is null limit 1","bindings":[9123]},"tags":[],"debugTrace":[{"file":"\/home\/www-data\/backend-rent\/releases\/130856_670596\/vendor\/laravel\/framework\/src\/Illuminate\/Database\/Connection.php","line":624,"function":"runQueryCallback","class":"BelkaCar\\LaravelIntegration\\TraceMonitor\\Database\\TraceMonitorPostgisConnection","type":"->"},{"file":"\/home\/www-data\/backend-rent\/releases\/130856_670596\/vendor\/laravel\/framework\/src\/Illuminate\/Database\/Connection.php","line":333,"function":"run","class":"Illuminate\\Database\\Connection","type":"->"},{"file":"\/home\/www-data\/backend-rent\/releases\/130856_670596\/vendor\/laravel\/framework\/src\/Illuminate\/Database\/Query\/Builder.php","line":1719,"function":"select","class":"Illuminate\\Database\\Connection","type":"->"}]},"parentSpans":[]}}`,
		`{"method":"set-trace-current-span","sentAt":"2023-04-10T14:04:31.394753+03:00","pid":"2905890","traceId":"0bbf9e15-519d-4e4f-af14-eb4caa40e88b","data":null}`,
		`{"method":"set-trace-current-span","sentAt":"2023-04-10T14:04:31.396019+03:00","pid":"2905890","traceId":"0bbf9e15-519d-4e4f-af14-eb4caa40e88b","data":{"span":{"id":"b8cd84b4-2a04-4f04-a05c-f589bbb82b1e","parent":null,"openedAt":"2023-04-10T14:04:31.396012+03:00","name":"Redis command","context":{"commandID":"get","arguments":["rent.cache:car_address:9123"],"config":{"host":"rent-redis-01.belkacar.ru","password":null,"port":6379,"database":"1"},"formattedOptions":{"timeout":10}},"tags":[],"debugTrace":[{"file":"\/home\/www-data\/backend-rent\/releases\/130856_670596\/vendor\/laravel\/framework\/src\/Illuminate\/Redis\/Connections\/Connection.php","line":96,"function":"__call","class":"class@anonymous\u0000\/home\/www-data\/backend-rent\/releases\/130856_670596\/vendor\/belkacar\/laravel-integration\/src\/IntegrationServiceProvider.php:120$48","type":"->"},{"file":"\/home\/www-data\/backend-rent\/releases\/130856_670596\/vendor\/laravel\/framework\/src\/Illuminate\/Redis\/Connections\/Connection.php","line":108,"function":"command","class":"Illuminate\\Redis\\Connections\\Connection","type":"->"},{"file":"\/home\/www-data\/backend-rent\/releases\/130856_670596\/vendor\/laravel\/framework\/src\/Illuminate\/Cache\/RedisStore.php","line":54,"function":"__call","class":"Illuminate\\Redis\\Connections\\Connection","type":"->"}]},"parentSpans":[]}}`,
		`{"method":"set-trace-current-span","sentAt":"2023-04-10T14:04:31.397478+03:00","pid":"2905890","traceId":"0bbf9e15-519d-4e4f-af14-eb4caa40e88b","data":null}`,
		`{"method":"set-trace-current-span","sentAt":"2023-04-10T14:04:31.398114+03:00","pid":"2905890","traceId":"0bbf9e15-519d-4e4f-af14-eb4caa40e88b","data":{"span":{"id":"21ab7b17-c00a-4475-9567-1ec11f58e3ee","parent":null,"openedAt":"2023-04-10T14:04:31.398107+03:00","name":"Redis command","context":{"commandID":"exists","arguments":["car_main_data_id_9123"],"config":{"host":"rent-redis-01.belkacar.ru","password":null,"port":"6379","database":0},"formattedOptions":{"timeout":10}},"tags":[],"debugTrace":[{"file":"\/home\/www-data\/backend-rent\/releases\/130856_670596\/vendor\/laravel\/framework\/src\/Illuminate\/Redis\/Connections\/Connection.php","line":96,"function":"__call","class":"class@anonymous\u0000\/home\/www-data\/backend-rent\/releases\/130856_670596\/vendor\/belkacar\/laravel-integration\/src\/IntegrationServiceProvider.php:120$48","type":"->"},{"file":"\/home\/www-data\/backend-rent\/releases\/130856_670596\/vendor\/laravel\/framework\/src\/Illuminate\/Redis\/Connections\/Connection.php","line":108,"function":"command","class":"Illuminate\\Redis\\Connections\\Connection","type":"->"},{"file":"\/home\/www-data\/backend-rent\/releases\/130856_670596\/app\/Repositories\/CarMainData\/CarMainDataCacheRepository.php","line":58,"function":"__call","class":"Illuminate\\Redis\\Connections\\Connection","type":"->"}]},"parentSpans":[]}}`,
		`{"method":"set-trace-current-span","sentAt":"2023-04-10T14:04:31.399197+03:00","pid":"2905890","traceId":"0bbf9e15-519d-4e4f-af14-eb4caa40e88b","data":null}`,
		`{"method":"set-trace-current-span","sentAt":"2023-04-10T14:04:31.399593+03:00","pid":"2905890","traceId":"0bbf9e15-519d-4e4f-af14-eb4caa40e88b","data":{"span":{"id":"d62c402a-9389-4da7-9535-7d61e9765f7a","parent":null,"openedAt":"2023-04-10T14:04:31.399586+03:00","name":"Redis command","context":{"commandID":"get","arguments":["car_main_data_id_9123"],"config":{"host":"rent-redis-01.belkacar.ru","password":null,"port":"6379","database":0},"formattedOptions":{"timeout":10}},"tags":[],"debugTrace":[{"file":"\/home\/www-data\/backend-rent\/releases\/130856_670596\/vendor\/laravel\/framework\/src\/Illuminate\/Redis\/Connections\/Connection.php","line":96,"function":"__call","class":"class@anonymous\u0000\/home\/www-data\/backend-rent\/releases\/130856_670596\/vendor\/belkacar\/laravel-integration\/src\/IntegrationServiceProvider.php:120$48","type":"->"},{"file":"\/home\/www-data\/backend-rent\/releases\/130856_670596\/vendor\/laravel\/framework\/src\/Illuminate\/Redis\/Connections\/Connection.php","line":108,"function":"command","class":"Illuminate\\Redis\\Connections\\Connection","type":"->"},{"file":"\/home\/www-data\/backend-rent\/releases\/130856_670596\/app\/Repositories\/CarMainData\/CarMainDataCacheRepository.php","line":61,"function":"__call","class":"Illuminate\\Redis\\Connections\\Connection","type":"->"}]},"parentSpans":[]}}`,
		`{"method":"set-trace-current-span","sentAt":"2023-04-10T14:04:31.400258+03:00","pid":"2905890","traceId":"0bbf9e15-519d-4e4f-af14-eb4caa40e88b","data":null}`,
		`{"method":"set-trace-current-span","sentAt":"2023-04-10T14:04:31.401568+03:00","pid":"2905890","traceId":"0bbf9e15-519d-4e4f-af14-eb4caa40e88b","data":{"span":{"id":"c08b56dd-e53c-49d1-8c21-23e8161e214d","parent":null,"openedAt":"2023-04-10T14:04:31.401560+03:00","name":"Redis command","context":{"commandID":"exists","arguments":["car_telematics_data_imei_353465070482615"],"config":{"host":"rent-redis-01.belkacar.ru","password":null,"port":"6379","database":0},"formattedOptions":{"timeout":10}},"tags":[],"debugTrace":[{"file":"\/home\/www-data\/backend-rent\/releases\/130856_670596\/vendor\/laravel\/framework\/src\/Illuminate\/Redis\/Connections\/Connection.php","line":96,"function":"__call","class":"class@anonymous\u0000\/home\/www-data\/backend-rent\/releases\/130856_670596\/vendor\/belkacar\/laravel-integration\/src\/IntegrationServiceProvider.php:120$48","type":"->"},{"file":"\/home\/www-data\/backend-rent\/releases\/130856_670596\/vendor\/laravel\/framework\/src\/Illuminate\/Redis\/Connections\/Connection.php","line":108,"function":"command","class":"Illuminate\\Redis\\Connections\\Connection","type":"->"},{"file":"\/home\/www-data\/backend-rent\/releases\/130856_670596\/app\/Repositories\/CarTelematicsData\/TelematicsDataHotRepository.php","line":156,"function":"__call","class":"Illuminate\\Redis\\Connections\\Connection","type":"->"}]},"parentSpans":[]}}`,
		`{"method":"set-trace-current-span","sentAt":"2023-04-10T14:04:31.401903+03:00","pid":"2905890","traceId":"0bbf9e15-519d-4e4f-af14-eb4caa40e88b","data":null}`,
		`{"method":"set-trace-current-span","sentAt":"2023-04-10T14:04:31.402205+03:00","pid":"2905890","traceId":"0bbf9e15-519d-4e4f-af14-eb4caa40e88b","data":{"span":{"id":"f973a816-5aee-4a37-b9d5-ea40f8975bcb","parent":null,"openedAt":"2023-04-10T14:04:31.402199+03:00","name":"Redis command","context":{"commandID":"get","arguments":["car_telematics_data_imei_353465070482615"],"config":{"host":"rent-redis-01.belkacar.ru","password":null,"port":"6379","database":0},"formattedOptions":{"timeout":10}},"tags":[],"debugTrace":[{"file":"\/home\/www-data\/backend-rent\/releases\/130856_670596\/vendor\/laravel\/framework\/src\/Illuminate\/Redis\/Connections\/Connection.php","line":96,"function":"__call","class":"class@anonymous\u0000\/home\/www-data\/backend-rent\/releases\/130856_670596\/vendor\/belkacar\/laravel-integration\/src\/IntegrationServiceProvider.php:120$48","type":"->"},{"file":"\/home\/www-data\/backend-rent\/releases\/130856_670596\/vendor\/laravel\/framework\/src\/Illuminate\/Redis\/Connections\/Connection.php","line":108,"function":"command","class":"Illuminate\\Redis\\Connections\\Connection","type":"->"},{"file":"\/home\/www-data\/backend-rent\/releases\/130856_670596\/app\/Repositories\/CarTelematicsData\/TelematicsDataHotRepository.php","line":159,"function":"__call","class":"Illuminate\\Redis\\Connections\\Connection","type":"->"}]},"parentSpans":[]}}`,
		`{"method":"set-trace-current-span","sentAt":"2023-04-10T14:04:31.402632+03:00","pid":"2905890","traceId":"0bbf9e15-519d-4e4f-af14-eb4caa40e88b","data":null}`,
		`{"method":"set-trace-current-span","sentAt":"2023-04-10T14:04:31.458611+03:00","pid":"2905890","traceId":"0bbf9e15-519d-4e4f-af14-eb4caa40e88b","data":{"span":{"id":"2459e234-de46-40db-9c85-12977dcd15a4","parent":null,"openedAt":"2023-04-10T14:04:31.458600+03:00","name":"Redis command","context":{"commandID":"setex","arguments":["rent.cache:car_address:9123",3600,"s:43:\"\u0443\u043b\u0438\u0446\u0430 \u041c\u0430\u0448\u0438 \u041f\u043e\u0440\u044b\u0432\u0430\u0435\u0432\u043e\u0439 34\";"],"config":{"host":"rent-redis-01.belkacar.ru","password":null,"port":6379,"database":"1"},"formattedOptions":{"timeout":10}},"tags":[],"debugTrace":[{"file":"\/home\/www-data\/backend-rent\/releases\/130856_670596\/vendor\/laravel\/framework\/src\/Illuminate\/Redis\/Connections\/Connection.php","line":96,"function":"__call","class":"class@anonymous\u0000\/home\/www-data\/backend-rent\/releases\/130856_670596\/vendor\/belkacar\/laravel-integration\/src\/IntegrationServiceProvider.php:120$48","type":"->"},{"file":"\/home\/www-data\/backend-rent\/releases\/130856_670596\/vendor\/laravel\/framework\/src\/Illuminate\/Redis\/Connections\/Connection.php","line":108,"function":"command","class":"Illuminate\\Redis\\Connections\\Connection","type":"->"},{"file":"\/home\/www-data\/backend-rent\/releases\/130856_670596\/vendor\/laravel\/framework\/src\/Illuminate\/Cache\/RedisStore.php","line":93,"function":"__call","class":"Illuminate\\Redis\\Connections\\Connection","type":"->"}]},"parentSpans":[]}}`,
		`{"method":"set-trace-current-span","sentAt":"2023-04-10T14:04:31.459019+03:00","pid":"2905890","traceId":"0bbf9e15-519d-4e4f-af14-eb4caa40e88b","data":null}`,
		`{"method":"free-pid","sentAt":"2023-04-10T14:04:31.461236+03:00","pid":"2905890","traceId":"0bbf9e15-519d-4e4f-af14-eb4caa40e88b","data":null}`,
	}

	// Act
	for _, pkt := range packets {
		_, err = client.Write([]byte(pkt))
		if err != nil {
			break
		}
	}

	time.Sleep(time.Second)

	// Assert
	assert.Nil(t, err)
	assert.Equal(t, 28, int(totalPackagesParse.Count()))
	assert.Equal(t, 28, int(totalPackagesCaught.Count()))
	assert.Equal(t, 1, int(traceCollection.TotalTraceSet.Count()))
	assert.Equal(t, 13, int(traceCollection.TotalSpanSet.Count()))
	assert.Equal(t, 0, int(traceCollection.CountActivePid.Count()))
	assert.Equal(t, 1, int(traceCollection.TotalTraceDelete.Count()))
	assert.Equal(t, 13, int(traceCollection.TotalAllSpanClose.Count()))

	cancel()
}

func TestNoTraceCollectedWhenSpanSentOutsideOfTrace(t *testing.T) {
	// Arrange
	ctx, cancel := context.WithCancel(context.Background())
	go handleUdp(ctx, udpServerConfig(8081))
	<-udpServerReadyChan

	udpServer, err := net.ResolveUDPAddr("udp", ":8081")
	require.Nil(t, err)
	client, err := net.DialUDP("udp", nil, udpServer)
	require.Nil(t, err)
	defer client.Close()

	packets := []string{
		`{"method":"set-trace-current-span","sentAt":"2023-04-10T14:04:31.376382+03:00","pid":"2905890","traceId":"0bbf9e15-519d-4e4f-af14-eb4caa40e88b","data":null}`,
	}

	// Act
	for _, pkt := range packets {
		_, err = client.Write([]byte(pkt))
		if err != nil {
			break
		}
	}

	time.Sleep(time.Second)

	// Assert
	assert.Nil(t, err)
	trace := traceCollection.GetAllTrace()
	assert.Empty(t, trace)

	cancel()
}

func udpServerConfig(port int) *config.Config {
	return &config.Config{
		UdpPortStart:         port,
		UdpPortEnd:           port,
		UdpPortRangeCount:    1,
		PacketsSize:          100,
		Buffer:               1048576,
		StuckProcessDuration: 10,
		LoadFpmStatusTimeout: 10,
		HttpClientTimeout:    3,
	}
}
