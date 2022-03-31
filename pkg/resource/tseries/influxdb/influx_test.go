package influxdb

// func Test_Write(t *testing.T) {
// outer := newInflux()
// outer.Init(resource.Metadata{
// 	Name: "influxdb",
// 	Properties: map[string]string{
// 		"org":    "yunify",
// 		"bucket": "entity",
// 		"url":    "http://localhost:8086",
// 		"token":  "9bUWcVwUpxbNSuhJMLbRaJxCVl8LzFV33znGx-pAXg4HUxFgWRTkRArF5Z9lMDcOn1pzzfD4dovLkkTnxuVMtg==",
// 	},
// })

// num := 10000
// t.Log("write som data, N=", num)
// for i := 0; i < num; i++ {
// 	_, err := outer.Write(context.Background(), &tseries.TSeriesRequest{
// 		Data: []string{fmt.Sprintf("mem,host=host1 used_percent=%f %d", 40.0, time.Now().Unix())},
// 	})
// 	if nil != err {
// 		t.Log("write influx failed", err)
// 	}
// }
// }
