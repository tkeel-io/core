import json
import random
import time
import traceback
import uuid

import requests
from paho.mqtt import client as mqtt_client

keel_url = "http://192.168.123.9:30707/v0.1.0"
broker = "192.168.123.9"
port = 32412


def create_entity_token(entity_id, entity_type, user_id):
    data = dict(entity_id=entity_id, entity_type=entity_type, user_id=user_id)

    token_create = "/auth/token/create"

    res = requests.post(keel_url + token_create, json=data)
    return res.json()["data"]["entity_token"]


def create_entity(entity_id, entity_type, user_id, plugin_id, token):
    query = dict(entity_id=entity_id, entity_type=entity_type, user_id=user_id, source="abc", plugin_id=plugin_id)
    entity_create = "/core/plugins/{plugin_id}/entities?id={entity_id}&type={entity_type}&owner={user_id}&source={source}".format(
        **query)
    data = dict(token=token)
    res = requests.post(keel_url + entity_create, json=data)
    print(res.json())


def create_subscription(entity_id, entity_type, user_id, plugin_id, subscription_id):
    query = dict(entity_id=entity_id, entity_type=entity_type, user_id=user_id, source="abc", plugin_id=plugin_id, subscription_id=subscription_id)
    entity_create = "/core/plugins/{plugin_id}/subscriptions?id={subscription_id}&type={entity_type}&owner={user_id}&source={source}".format(
        **query)
    data = dict(mode="realtime", source="ignore", filter="insert into abc select " + entity_id + ".p1", target="ignore", topic="abc", pubsub_name="client-pubsub")
    print(data)
    res = requests.post(keel_url + entity_create, json=data)
    print(res.json())


def get_subscription(entity_id, entity_type, user_id, plugin_id, subscription_id):
    query = dict(entity_id=entity_id, entity_type=entity_type, user_id=user_id, source="abc", plugin_id=plugin_id, subscription_id=subscription_id)
    entity_create = "/core/plugins/{plugin_id}/subscriptions/{subscription_id}?type={entity_type}&owner={user_id}&source={source}".format(
        **query)
    res = requests.get(keel_url + entity_create)
    print(res.json())


def get_entity(entity_id, entity_type, user_id, plugin_id):
    query = dict(entity_id=entity_id, entity_type=entity_type, user_id=user_id, plugin_id=plugin_id)
    entity_create = "/core/plugins/{plugin_id}/entities/{entity_id}?type={entity_type}&owner={user_id}&source={plugin_id}".format(
        **query)
    res = requests.get(keel_url + entity_create)
    print(res.json()["properties"])


def on_connect(client, userdata, flags, rc):
    if rc == 0:
        print("Connected to MQTT Broker!")
    else:
        print("Failed to connect, return code %d\n", rc)


if __name__ == "__main__":
    entity_id = uuid.uuid4().hex
    entity_type = "device"
    user_id = "abc"
    print("base entity info")
    print("entity_id = ", entity_id)
    print("entity_type = ", entity_type)
    print("user_id = ", user_id)

    print("-" * 80)
    print("get entity token")
    token = create_entity_token(entity_id, entity_type, user_id)
    print("token=", token)
    time.sleep(1)
    print("-" * 80)
    print("create entity with token")
    try:
        create_entity(entity_id, entity_type, user_id, "pluginA", token)
        print("create entity {entity_id} success".format(**dict(entity_id=entity_id)))
    except Exception:
        print(traceback.format_exc())
        print("create entity failed")
    time.sleep(1)

    print("-" * 80)
    print("create subscription")
    create_subscription(entity_id, "SUBSCRIPTION", user_id, "pluginA", entity_id+"sub")
    print("-" * 80)
    print("get subscription")
    get_subscription(entity_id, "SUBSCRIPTION", user_id, "pluginA", entity_id+"sub")
    print("-" * 80)
    print("update properties by mqtt")
    client = mqtt_client.Client(entity_id)

    client.username_pw_set(username=user_id, password=token)
    client.on_connect = on_connect
    client.connect(host=broker, port=port)
    client.loop_start()
    time.sleep(1)
    payload = json.dumps(dict(p1=dict(value=random.randint(1, 100), time=int(time.time()))))
    print(payload)
    client.publish("system/test", payload=payload)
    print("-" * 80)
    print("get entity")
    get_entity(entity_id, entity_type, user_id, "pluginA")
    time.sleep(5)
    while True:
        payload = json.dumps(dict(p1=dict(value=random.randint(1, 100), time=int(time.time()))))
        print(payload)
        client.publish("system/test", payload=payload)
        time.sleep(5)
    client.disconnect()
