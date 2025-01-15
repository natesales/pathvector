"""
PeeringDB test API

Supported queries:
https://peeringdb.com/api/net?info_never_via_route_servers=1
https://peeringdb.com/api/net?asn=%d
https://peeringdb.com/api/netixlan?asn=%d
"""

import json

from flask import Flask, jsonify, request

app = Flask(__name__)

nvrs = json.loads(open("nvrs.json").read())
nets = json.loads(open("net.json").read())
netixlans = json.loads(open("netixlan.json").read())


def response(d):
    return {"data": [d], "meta": {}}

@app.route("/")
def index():
    return "PeeringDB test API"

@app.route("/api/net", methods=["GET"])
def net():
    if request.args.get("info_never_via_route_servers") == "1":
        return jsonify(nvrs)

    asn = int(request.args.get("asn"))
    for net in nets["data"]:
        if net["asn"] == asn:
            return jsonify(response(net))
    return jsonify({"data": [], "meta": {"error": "Entity not found"}})


@app.route("/api/netixlan", methods=["GET"])
def netixlan():
    asn = int(request.args.get("asn"))
    for net in netixlans["data"]:
        if net["asn"] == asn:
            return jsonify(response(net))
    return jsonify({"data": [], "meta": {"error": "Entity not found"}})


if __name__ == "__main__":
    app.run("0.0.0.0", port=5001)
