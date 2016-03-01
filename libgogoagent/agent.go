package libgogoagent

import (
	"fmt"
	"golang.org/x/net/websocket"
	"log"
	"os"
	"runtime"
	"time"
)

var (
	uuid          = "564e9408-fb78-4856-4215-52e0-e14bb056"
	serverHost    = "localhost"
	sslPort       = "8154"
	httpPort      = "8153"
	hostname, _   = os.Hostname()
	workingDir, _ = os.Getwd()
)

func sslHostAndPort() string {
	return serverHost + ":" + sslPort
}

func httpsServerURL(path string) string {
	return "https://" + sslHostAndPort() + path
}

func httpServerURL(path string) string {
	return "http://" + serverHost + ":" + httpPort + path
}

func StartAgent() {
	ReadGoServerCACert()
	Register(map[string]string{
		"hostname":                      hostname,
		"uuid":                          uuid,
		"location":                      workingDir,
		"operatingSystem":               runtime.GOOS,
		"usablespace":                   "5000000000",
		"agentAutoRegisterKey":          "",
		"agentAutoRegisterResources":    "",
		"agentAutoRegisterEnvironments": "",
		"agentAutoRegisterHostname":     "",
		"elasticAgentId":                "",
		"elasticPluginId":               "",
	})

	loc := "wss://" + GoServerDN() + ":8154/go/agent-websocket"
	config, _ := websocket.NewConfig(loc, httpsServerURL("/"))
	config.TlsConfig = GoServerTlsConfig(true)
	ws, err := websocket.DialConfig(config)
	if err != nil {
		panic(err)
	}

	buildSession := BuildSession{
		HttpClient: GoServerRemoteClient(true)}

	go ping(ws)

	for {
		var msg Message
		err := MessageCodec.Receive(ws, &msg)
		if err != nil {
			log.Println("Can't decode message received")
		}
		switch msg.Action {
		case "setCookie":
			str, _ := msg.Data["data"].(string)
			SetState("cookie", str)
		case "cmd":
			err = processCmdMessage(&buildSession, &msg)
		}
		if err != nil {
			log.Println(fmt.Sprintf("Error(%v) when processing message : %v", err, msg))
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func processCmdMessage(buildSession *BuildSession, msg *Message) error {
	SetState("runtimeStatus", "Building")
	defer SetState("runtimeStatus", "Idle")

	command, _ := msg.Data["data"].(map[string]interface{})
	return buildSession.Process(MakeBuildCommand(command))
}

func ping(ws *websocket.Conn) {
	for {
		data := make(map[string]interface{})
		data["identifier"] = map[string]string{
			"hostName":  hostname,
			"ipAddress": "127.0.0.1",
			"uuid":      uuid}
		data["runtimeStatus"] = GetState("runtimeStatus")
		data["buildingInfo"] = map[string]string{
			"buildingInfo": GetState("buildingInfo"),
			"buildLocator": GetState("buildLocator")}
		data["location"] = workingDir
		data["usableSpace"] = "12262604800"
		data["operatingSystemName"] = runtime.GOOS
		data["agentLauncherVersion"] = ""

		if cookie := GetState("cookie"); cookie != "" {
			data["cookie"] = cookie
		}

		msg := Message{"ping", map[string]interface{}{
			"type": "com.thoughtworks.go.server.service.AgentRuntimeInfo",
			"data": data}}

		MessageCodec.Send(ws, msg)
		time.Sleep(10 * time.Second)
	}
}
