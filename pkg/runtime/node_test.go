package runtime

import (
	"context"
	"encoding/json"
	"net/url"
	"testing"
)

// func TestNode_Start(t *testing.T) {
// 	stopCh := make(chan struct{})
// 	placement.Initialize()
// 	log.InitLogger("core.node", "DEBUG", true)
// 	node := NewNode(context.Background(), nil, mock.NewDispatcher())

// 	err := node.Start(NodeConf{
// 		Sources: []string{
// 			"kafka://139.198.125.147:9092/core/core",
// 		}})

// 	if nil != err {
// 		panic(err)
// 	}

// 	<-stopCh
// }

func TestParse(t *testing.T) {
	urlText := "partition://admin:admin@192.168.12.1;192.168.12.1/core/0"
	URL, _ := url.Parse(urlText)
	t.Log(URL)
}

func TestNode_getGlobalData(t *testing.T) {
	node := NewNode(context.Background(), nil, nil)

	entityBytes := `{
        "id": "iotd-a4375b93-a9fd-417c-b6a4-5ec8ecb87f41",
        "source": "device",
        "owner": "usr-3358ac43d4ca8a05fee8a6db7b14",
        "type": "device",
        "version": "13",
        "last_time": "1649824136703",
        "template_id": "",
        "description": "",
        "mappers": [
            {
                "id": "mapper_space_path",
                "name": "mapper_space_path",
                "tql": "insert into iotd-a4375b93-a9fd-417c-b6a4-5ec8ecb87f41 select iotd-d2ac161d-ebff-4239-ac32-de0f277075b7.properties.sysField._spacePath + '/iotd-a4375b93-a9fd-417c-b6a4-5ec8ecb87f41'  as properties.sysField._spacePath",
                "description": ""
            }
        ],
        "configs": {},
        "properties": {
            "basicInfo": {
                "description": "",
                "directConnection": true,
                "ext": {
                    "安装时间": "123",
                    "负责人": "dddd"
                },
                "name": "lllllllll",
                "parentId": "iotd-d2ac161d-ebff-4239-ac32-de0f277075b7",
                "parentName": "",
                "selfLearn": false,
                "templateId": "",
                "templateName": ""
            },
            "connectInfo": {
                "_clientId": "",
                "_online": false,
                "_peerHost": "",
                "_protocol": "",
                "_sockPort": "",
                "_userName": ""
            },
            "sysField1": {
                "_createdAt": 1649820366777,
                "_enable": true,
                "_id": "iotd-a4375b93-a9fd-417c-b6a4-5ec8ecb87f41",
                "_owner": "usr-3358ac43d4ca8a05fee8a6db7b14",
                "_source": "device",
                "_spacePath": "iotd-usr-3358ac43d4ca8a05fee8a6db7b14-defaultGroup/iotd-d2ac161d-ebff-4239-ac32-de0f277075b7/iotd-a4375b93-a9fd-417c-b6a4-5ec8ecb87f41",
                "_status": "offline",
                "_subscribeAddr": "",
                "_token": "NGM0M2ViOTQtMzNlOC0zOGE2LTk2YTQtZTU4NjJjOTFmMDMz",
                "_updatedAt": 1649824132030
            }
        }
    }`
	en, err := NewEntity("iotd-a4375b93-a9fd-417c-b6a4-5ec8ecb87f41", []byte(entityBytes))
	if err != nil {
		t.Log(err)
	}
	t.Log(en)
	res := node.getGlobalData(en)
	t.Log(string(res))

	resMap := make(map[string]interface{})

	err = json.Unmarshal(res, &resMap)
	t.Log(err)
}




func Test_parseExpression(t *testing.T) {
    
}