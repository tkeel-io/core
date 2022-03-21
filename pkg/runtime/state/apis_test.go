package state

import "testing"

func TestMakesubPath(t *testing.T) {
	dest := []byte(` {
        "tempsensor": {
            "id": "tempsensor",
            "type": "struct",
            "weight": 0
            "last_time": 0,
            "description": "",
            "enabled": true,
            "enabled_search": false,
            "enabled_time_series": false,
            "define": {
                "fields": [
                    {
                        "define": {
                            "max": 500,
                            "min": 10,
                            "unit": "°"
                        },
                        "description": "",
                        "enabled": true,
                        "enabled_search": false,
                        "enabled_time_series": false,
                        "id": "temp",
                        "last_time": 0,
                        "type": "int",
                        "weight": 0
                    }
                ]
            }
        }
    }`)

	src := []byte(` {
		"define": {
			"max": 500,
			"min": 10,
			"unit": "°"
		},
		"description": "",
		"enabled": true,
		"enabled_search": false,
		"enabled_time_series": false,
		"id": "temp2",
		"last_time": 0,
		"type": "int",
		"weight": 0
	}`)

	res, err := makeSubPath(dest, src, "tempsensor.x.temp2")
	t.Log(err)
	t.Log(string(res))
}
