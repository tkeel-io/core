package state

import "testing"

func TestMakesubPath(t *testing.T) {
	dest := []byte(`{
        "telemetry": {
            "aaa": {
                "define": {
                    "ext": {
                        "test": 123
                    },
                    "max": 1,
                    "min": 0
                },
                "description": "",
                "enabled": true,
                "enabled_search": true,
                "enabled_time_series": false,
                "id": "",
                "last_time": 0,
                "name": "",
                "type": "float",
                "weight": 0
            },
            "ddd": {
                "define": {
                    "ext": {
                        "test": 123
                    },
                    "max": 1,
                    "min": 0
                },
                "description": "",
                "enabled": true,
                "enabled_search": true,
                "enabled_time_series": false,
                "id": "",
                "last_time": 0,
                "name": "",
                "type": "float",
                "weight": 0
            },
            "define": {
                "ext": {
                    "test": 123
                },
                "fields": {
                    "aaa": {
                        "define": {
                            "ext": {
                                "test": 123
                            },
                            "max": 1,
                            "min": 0
                        },
                        "description": "",
                        "enabled": true,
                        "enabled_search": true,
                        "enabled_time_series": false,
                        "id": "",
                        "last_time": 0,
                        "name": "",
                        "type": "float",
                        "weight": 0
                    },
                    "bbb": {
                        "define": {
                            "ext": {
                                "test": 123
                            },
                            "max": 1,
                            "min": 0
                        },
                        "description": "",
                        "enabled": true,
                        "enabled_search": true,
                        "enabled_time_series": false,
                        "id": "",
                        "last_time": 0,
                        "name": "",
                        "type": "float",
                        "weight": 0
                    },
                    "ddd": {
                        "define": {
                            "ext": {
                                "test": 123
                            },
                            "max": 1,
                            "min": 0
                        },
                        "description": "",
                        "enabled": true,
                        "enabled_search": true,
                        "enabled_time_series": false,
                        "id": "",
                        "last_time": 0,
                        "name": "",
                        "type": "float",
                        "weight": 0
                    },
                    "vvv": {
                        "define": {
                            "ext": {
                                "test": 123
                            },
                            "max": 1,
                            "min": 0
                        },
                        "description": "",
                        "enabled": true,
                        "enabled_search": true,
                        "enabled_time_series": false,
                        "id": "",
                        "last_time": 0,
                        "name": "",
                        "type": "float",
                        "weight": 0
                    },
                    "xx3": {
                        "define": {
                            "fields": {
                                "define": {
                                    "define": {
                                        "fields": {
                                            "nnn": {
                                                "define": {
                                                    "ext": {
                                                        "test": 123
                                                    },
                                                    "max": 1,
                                                    "min": 0
                                                },
                                                "description": "",
                                                "enabled": true,
                                                "enabled_search": true,
                                                "enabled_time_series": false,
                                                "id": "",
                                                "last_time": 0,
                                                "name": "",
                                                "type": "float",
                                                "weight": 0
                                            }
                                        }
                                    },
                                    "description": "",
                                    "enabled": true,
                                    "enabled_search": true,
                                    "enabled_time_series": true,
                                    "id": "define",
                                    "last_time": 1647853077820,
                                    "name": "define",
                                    "type": "struct",
                                    "weight": 0
                                }
                            }
                        },
                        "description": "",
                        "enabled": true,
                        "enabled_search": true,
                        "enabled_time_series": true,
                        "id": "xx3",
                        "last_time": 1647853077820,
                        "name": "xx3",
                        "type": "struct",
                        "weight": 0
                    },
                    "xxx3": {
                        "define": {
                            "fields": {
                                "define": {
                                    "define": {
                                        "fields": {
                                            "nnn": {
                                                "define": {
                                                    "ext": {
                                                        "test": 123
                                                    },
                                                    "max": 1,
                                                    "min": 0
                                                },
                                                "description": "",
                                                "enabled": true,
                                                "enabled_search": true,
                                                "enabled_time_series": false,
                                                "id": "",
                                                "last_time": 0,
                                                "name": "",
                                                "type": "float",
                                                "weight": 0
                                            }
                                        }
                                    },
                                    "description": "",
                                    "enabled": true,
                                    "enabled_search": true,
                                    "enabled_time_series": true,
                                    "id": "define",
                                    "last_time": 1647853228985,
                                    "name": "define",
                                    "type": "struct",
                                    "weight": 0
                                }
                            }
                        },
                        "description": "",
                        "enabled": true,
                        "enabled_search": true,
                        "enabled_time_series": true,
                        "id": "xxx3",
                        "last_time": 1647853228985,
                        "name": "xxx3",
                        "type": "struct",
                        "weight": 0
                    }
                },
                "max": 1,
                "min": 0
            },
            "description": "",
            "enabled": true,
            "enabled_search": true,
            "enabled_time_series": false,
            "id": "",
            "last_time": 0,
            "name": "",
            "type": "float",
            "weight": 0
        }
    }`)

	src := []byte(` {
        "type": "float",
        "define": {
            "max": 1,
            "min": 0,
            "ext": {
                "test": 123
            }
        },
        "enabled": true,
        "enabled_search": true
    }`)

	res, path, err := makeSubPath(dest, src, "telemetry.define.fields.x.define.fields.temp2")
	t.Log(err)
	t.Log(path)
	t.Log(string(res))
}
