package main

import (
	"fmt"
	"sync"
	"net/http"
	"encoding/json"
	"os/exec"
	"time"
	"bytes"
	"os"
	"strings"
)

type VethPair struct {
	SwitchL string `json:"switchL"`
	SwitchR string `json:"switchR"`
}
type Switch struct {
	SwitchName           string `json:"switchName"`
	SwitchControllerIp   string `json:"switchControllerIp"`
	SwitchControllerPort string `json:"switchControllerPort"`
}

type Topology struct {
	Switches  []Switch
	VethPairs []VethPair
}

type VnfDocker struct {
	Name    string `json:"name"`
	Image   string `json:"img"`
	Cpu     string `json:"cpu"`
	Ram     string `json:"ram"`
	Command string `json:"command"`
}

type NetworkCard struct {
	Name       string `json:"name"`
	IpAddress  string `json:"ipAddress"`
	MacAddress string `json:"macAddress"`
}

type HostStatus struct {
	Cpu          string `json:"cpu"`
	Ram          string `json:"ram"`
	Hhd          string `json:"hhd"`
	CpuUsage     string `json:"cpuUsage"`
	RamUsage     string `json:"ramUsage"`
	HhdUsage     string `json:"hhdUsage"`
	OvsVersion   string `json:"ovsVersion"`
}

type OVSDockerPort struct {
	VnfName          string `json:"vnfName"`
	VnfIpAddress     string `json:"vnfIpAddress"`
	VnfInterfaceName string `json:"vnfInterfaceName"`
	SwitchName       string `json:"switchName"`
}

type SflowAgent struct {
	SwitchName      string `json:"switchName"`
	AgentId         string `json:"agentId"`
	SenderInterface string `json:"senderInterface"`
	CollectorIp     string `json:"collectorIp"`
	CollectorPort   string `json:"collectorPort"`
	SamplingRate    string `json:"samplingRate"`
	PollingRate     string `json:"pollingRate"`
}

type ContainerCommand struct {
	ContainerName string `json:"containerName"`
	CommandName   string `json:"commandName"`
	Command       string `json:"command"`
}

type Container struct {
	Image         string `json:"image"`
	ContainerName string `json:"containerName"`
	InitCommand   string `json:"initCommand"`
	Cpu           string `json:"cpu"`
	Ram           string `json:"ram"`
	Ports         string `json:"ports"`
}

func getHostStatus(wg *sync.WaitGroup) HostStatus {
	hostStatus := HostStatus{}

	// Get Cpu usage
	getCpuStatusCommand := "top -bn1 | grep 'Cpu(s)' | sed 's/.*, *\\([0-9.]*\\)%* id.*/\\1/' | awk '{print 100 - $1''}'"
	cpuStatus, errGetCpuStatusCommand := exec.Command("bash", "-c", getCpuStatusCommand).Output()
	if errGetCpuStatusCommand != nil {
		fmt.Printf("%s", errGetCpuStatusCommand)
		hostStatus.CpuUsage = ""
	} else {
		/*
		for preventing "\n" effect in marshalize json
		ٖٖ{
			"cpu": "2\n",
			"ram": "5\n",
			"hhd": "94.3687\n",
			"cpuUsage": "0.7\n",
			"ramUsage": "2.64248\n",
			"hhdUsage": "94.3687\n",
			"ovsVersion": ""
		}
		 */
		hostStatus.CpuUsage = strings.TrimSuffix(string(cpuStatus), "\n")
	}

	// Get ram usage (%)
	getRamStatusCommand := "free | grep Mem | awk '{print $3/$2 * 100.0}'"
	ramStatus, errGetRamStatusCommand := exec.Command("bash", "-c", getRamStatusCommand).Output()
	if errGetRamStatusCommand != nil {
		fmt.Printf("%s", errGetRamStatusCommand)
		hostStatus.RamUsage = ""
	} else {
		hostStatus.RamUsage = strings.TrimSuffix(string(ramStatus), "\n")
	}

	// Get hhd usage
	getHHDStatusCommand := "df -P / | awk '/%/ {print 100 -$5 ''}'"
	hhdStatus, errgetHHDStatusCommand := exec.Command("bash", "-c", getHHDStatusCommand).Output()
	if errgetHHDStatusCommand != nil {
		fmt.Printf("%s", errgetHHDStatusCommand)
		hostStatus.HhdUsage = ""
	} else {
		hostStatus.HhdUsage = strings.TrimSuffix(string(hhdStatus), "\n")
	}

	// Get Total Cpu
	getTotalCpuStatusCommand := "echo  $(( $(lscpu | awk '/^Socket/{ print $2 }') * $(lscpu | awk '/^Core/{ print $4 }') ))"
	totalCpu, errgetTotalCpuStatusCommand := exec.Command("bash", "-c", getTotalCpuStatusCommand).Output()
	if errgetHHDStatusCommand != nil {
		fmt.Printf("%s", errgetTotalCpuStatusCommand)
		hostStatus.Cpu = ""
	} else {
		hostStatus.Cpu = strings.TrimSuffix(string(totalCpu), "\n")
	}

	// Get Total ram (GB)
	getTotalRamStatusCommand := "expr $(awk '/MemTotal/ {print $2}' /proc/meminfo) / 1048576"
	totalRam, errgetTotalRamStatusCommand := exec.Command("bash", "-c", getTotalRamStatusCommand).Output()
	if errgetHHDStatusCommand != nil {
		fmt.Printf("%s", errgetTotalRamStatusCommand)
		hostStatus.Ram = ""
	} else {
		hostStatus.Ram = strings.TrimSuffix(string(totalRam), "\n")
	}
	// Get Total hhd
	getTotalHhdStatusCommand := "df | grep '^/dev/[hs]d' | awk '{s+=$2} END {print s/1048576}'"
	totalHhd, errgetTotalHhdStatusCommand := exec.Command("bash", "-c", getTotalHhdStatusCommand).Output()
	if errgetHHDStatusCommand != nil {
		fmt.Printf("%s", errgetTotalHhdStatusCommand)
		hostStatus.Hhd = ""
	} else {
		hostStatus.Hhd = strings.TrimSuffix(string(totalHhd), "\n")
	}

	wg.Done() // Need to signal to waitgroup that this goroutine is done
	t := time.Now()
	fmt.Println(t.Format("2006-01-02 15:04:05") + " --- " + "Get Host Status")
	return hostStatus
}

func getHostNetworkCardPropertyByName(networkCard NetworkCard, wg *sync.WaitGroup) NetworkCard {
	tempNetworkCard := NetworkCard{}
	tempNetworkCard.Name = networkCard.Name
	// Create Switch
	getNetworkCardIpAdd := "echo $(ifconfig | grep -A 1 '" + networkCard.Name + "' | tail -1 | cut -d ':' -f 2 | cut -d ' ' -f 1)"
	ipAddress, errgetNetworkCardIpAdd := exec.Command("bash", "-c", getNetworkCardIpAdd).Output()
	if errgetNetworkCardIpAdd != nil {
		fmt.Printf("%s", errgetNetworkCardIpAdd)
		tempNetworkCard.IpAddress = ""
	} else {
		tempNetworkCard.IpAddress = strings.TrimSuffix(string(ipAddress), "\n")
	}

	getNetworkCardMacAdd := "ifconfig " + networkCard.Name + " | awk '/HWaddr/ {print $5}'"
	macAddress, errgetNetworkCardMacAdd := exec.Command("bash", "-c", getNetworkCardMacAdd).Output()
	if errgetNetworkCardIpAdd != nil {
		fmt.Printf("%s", errgetNetworkCardMacAdd)
		tempNetworkCard.MacAddress = ""
	} else {
		tempNetworkCard.MacAddress = strings.TrimSuffix(string(macAddress), "\n")
	}

	wg.Done() // Need to signal to waitgroup that this goroutine is done
	t := time.Now()
	fmt.Println(t.Format("2006-01-02 15:04:05") + " --- " + "Got Host network cards")
	return tempNetworkCard
}

func createSwitch(switchName string, switchControllerIp string, switchControllerPort string, wg *sync.WaitGroup) string {
	// Create Switch
	createSwitch := "ovs-vsctl add-br " + switchName
	_, errCreateSwitch := exec.Command("bash", "-c", createSwitch).Output()
	if errCreateSwitch != nil {
		fmt.Printf("%s", errCreateSwitch)
	}

	// Set Controller
	if len(switchControllerIp) != 0 {
		if len(switchControllerPort) != 0 {
			setSwitchController(switchName, switchControllerIp, switchControllerPort, wg)
		}
		setSwitchController(switchName, switchControllerIp, "6633", wg)
	}

	wg.Done() // Need to signal to waitgroup that this goroutine is done
	//return string(outCreateIpLink[:])
	t := time.Now()
	fmt.Println(t.Format("2006-01-02 15:04:05") + " --- " + "Switch " + switchName + " Created")
	return "Switch " + switchName + " Created"
}

func deleteSwitch(switchName string, wg *sync.WaitGroup) string {
	// Create Switch
	delSwitch := "ovs-vsctl del-br " + switchName
	_, errDelSwitch := exec.Command("bash", "-c", delSwitch).Output()
	if errDelSwitch != nil {
		fmt.Printf("%s", errDelSwitch)
	}

	wg.Done() // Need to signal to waitgroup that this goroutine is done
	t := time.Now()
	fmt.Println(t.Format("2006-01-02 15:04:05") + " --- " + "Switch " + switchName + " Deleted")
	return "Switch " + switchName + " Deleted"
}

func setSwitchController(switchName string, switchControllerIp string, switchControllerPort string, wg *sync.WaitGroup) string {
	// Set Controller
	setController := "ovs-vsctl set-controller " + switchName + " tcp:" + switchControllerIp + ":" + switchControllerPort
	_, errSetController := exec.Command("bash", "-c", setController).Output()
	if errSetController != nil {
		fmt.Printf("%s", errSetController)
	}

	wg.Done() // Need to signal to waitgroup that this goroutine is done
	//return string(outCreateIpLink[:])
	t := time.Now()
	fmt.Println(t.Format("2006-01-02 15:04:05") + " --- " + "Switch " + switchName + " Controller set to " + " tcp:" + switchControllerIp + ":" + switchControllerPort)
	return "Switch " + switchName + " Controller set to " + " tcp:" + switchControllerIp + ":" + switchControllerPort
}

func createVethPair(switchL string, switchR string, wg *sync.WaitGroup) string {

	// Create Ip Links
	createIpLink := "ip link add " + switchL + "_" + switchR + " type veth peer name " + switchR + "_" + switchL
	_, errcreateIpLink := exec.Command("bash", "-c", createIpLink).Output()
	if errcreateIpLink != nil {
		fmt.Printf("%s", errcreateIpLink)
	}

	// StartUp Interfaces
	startUpInrefaceL := "ifconfig " + switchL + "_" + switchR + " up"
	_, errStartUpInrefaceL := exec.Command("bash", "-c", startUpInrefaceL).Output()
	if errStartUpInrefaceL != nil {
		fmt.Printf("%s", errStartUpInrefaceL)
	}

	startUpInrefaceR := "ifconfig " + switchR + "_" + switchL + " up"
	_, errStartUpInrefaceR := exec.Command("bash", "-c", startUpInrefaceR).Output()
	if errStartUpInrefaceR != nil {
		fmt.Printf("%s", errStartUpInrefaceR)
	}

	// ADD Interfaces to OVS Bridges
	addInterfaceL := "ovs-vsctl add-port " + switchL + " " + switchL + "_" + switchR
	_, erraddInterfaceL := exec.Command("bash", "-c", addInterfaceL).Output()
	if erraddInterfaceL != nil {
		fmt.Printf("%s", erraddInterfaceL)
	}

	addInterfaceR := "ovs-vsctl add-port " + switchR + " " + switchR + "_" + switchL
	_, erraddInterfaceR := exec.Command("bash", "-c", addInterfaceR).Output()
	if erraddInterfaceR != nil {
		fmt.Printf("%s", erraddInterfaceR)
	}

	wg.Done() // Need to signal to waitgroup that this goroutine is done
	t := time.Now()
	fmt.Println(t.Format("2006-01-02 15:04:05") + " --- " + "Veth pair " + switchL + "_" + switchR + " Created")
	return "Veth pair " + switchL + "_" + switchR + " Created"
}

func deleteVethPair(switchL string, switchR string, wg *sync.WaitGroup) string {

	// delete OVS links
	delOVSLinkL := "ovs-vsctl --if-exists del-port " + switchL + " " + switchL + "_" + switchR
	_, errDelOVSLinkL := exec.Command("bash", "-c", delOVSLinkL).Output()
	if errDelOVSLinkL != nil {
		fmt.Printf("%s", errDelOVSLinkL)
	}

	delOVSLinkR := "ovs-vsctl --if-exists del-port " + switchR + " " + switchR + "_" + switchL
	_, errDelOVSLinkR := exec.Command("bash", "-c", delOVSLinkR).Output()
	if errDelOVSLinkR != nil {
		fmt.Printf("%s", errDelOVSLinkR)
	}

	// Delete veth pair Interfaces
	delVethInterface := "ip link delete " + switchL + "_" + switchR
	_, errDelVethInterface := exec.Command("bash", "-c", delVethInterface).Output()
	if errDelVethInterface != nil {
		fmt.Printf("%s", errDelVethInterface)
	}

	wg.Done() // Need to signal to waitgroup that this goroutine is done
	t := time.Now()
	fmt.Println(t.Format("2006-01-02 15:04:05") + " --- " + "Veth pair " + switchL + "_" + switchR + " Deleted")
	return "Veth pair " + switchL + "_" + switchR + " Deleted"
}

func createVnfDocker(vnfDocker VnfDocker, wg *sync.WaitGroup) string {
	createVnfDocker := "docker run -dit  --name " + vnfDocker.Name + " --net=none "
	if (vnfDocker.Ram != "") {
		createVnfDocker += "--memory=\"" + vnfDocker.Ram + "m\" "
	}
	if (vnfDocker.Cpu != "") {
		createVnfDocker += "--cpus=\"" + vnfDocker.Cpu + "\" "
	}
	createVnfDocker += " " + vnfDocker.Image + " " + vnfDocker.Command

	fmt.Println(createVnfDocker)

	_, errCreateVnfDocker := exec.Command("bash", "-c", createVnfDocker).Output()
	if errCreateVnfDocker != nil {
		fmt.Printf("%s", errCreateVnfDocker)
	}

	wg.Done() // Need to signal to waitgroup that this goroutine is done
	t := time.Now()
	fmt.Println(t.Format("2006-01-02 15:04:05") + " --- " + "VNF " + vnfDocker.Name + " Created")
	return "VNF " + vnfDocker.Name + " Created"
}

func deleteVnfDocker(vnfName string, wg *sync.WaitGroup) string {
	// delete Container
	deleteVnfDocker := "docker rm -f " + vnfName
	_, errDeleteVnfDocker := exec.Command("bash", "-c", deleteVnfDocker).Output()
	if errDeleteVnfDocker != nil {
		fmt.Printf("%s", errDeleteVnfDocker)
	}

	wg.Done() // Need to signal to waitgroup that this goroutine is done
	t := time.Now()
	fmt.Println(t.Format("2006-01-02 15:04:05") + " --- " + "VNF " + vnfName + " Deleted")
	return "VNF " + vnfName + " Deleted"
}

func createOVSDockerPort(vnfName string, vnfIpAddress string, vnfInterfaceName string, switchName string, wg *sync.WaitGroup) string {

	createOVSDockerPort := "ovs-docker add-port " + switchName + " " + vnfInterfaceName + " " + vnfName + " --ipaddress=" + vnfIpAddress + "/24 --mtu=1400"
	_, errCreateOVSDockerPort := exec.Command("bash", "-c", createOVSDockerPort).Output()
	if errCreateOVSDockerPort != nil {
		fmt.Printf("%s", errCreateOVSDockerPort)
	}

	wg.Done() // Need to signal to waitgroup that this goroutine is done
	t := time.Now()
	fmt.Println(t.Format("2006-01-02 15:04:05") + " --- " + "ovs Docker port Created")
	return "ovs Docker port Created"
}

func deleteAllOVSDockerPort(vnfName string, switchName string, wg *sync.WaitGroup) string {

	deleteOVSDockerPort := "ovs-docker del-ports " + switchName + " " + vnfName
	_, errDeleteOVSDockerPort := exec.Command("bash", "-c", deleteOVSDockerPort).Output()
	if errDeleteOVSDockerPort != nil {
		fmt.Printf("%s", errDeleteOVSDockerPort)
	}

	wg.Done() // Need to signal to waitgroup that this goroutine is done
	t := time.Now()
	fmt.Println(t.Format("2006-01-02 15:04:05") + " --- " + "ovs Docker port Deleted")
	return "ovs Docker port Deleted"
}

func deleteOVSDockerPort(vnfName string, inaterfaceName string, switchName string, wg *sync.WaitGroup) string {

	deleteOVSDockerPort := "ovs-docker del-port " + switchName + " " + inaterfaceName + " " + vnfName
	_, errDeleteOVSDockerPort := exec.Command("bash", "-c", deleteOVSDockerPort).Output()
	if errDeleteOVSDockerPort != nil {
		fmt.Printf("%s", errDeleteOVSDockerPort)
	}

	wg.Done() // Need to signal to waitgroup that this goroutine is done
	t := time.Now()
	fmt.Println(t.Format("2006-01-02 15:04:05") + " --- " + "ovs Docker port Deleted")
	return "ovs Docker port Deleted"
}

func setSflowAgent(switchName string, agentId string, senderInterface string, collectorIp string, collectorPort string, samplingRate string, pollingRate string, wg *sync.WaitGroup) string {

	setSflowAgentCommand := "ovs-vsctl -- --id=" + agentId + " create sflow agent=" + senderInterface + " target=\\\"" + collectorIp + ":" + collectorPort + "\\\" sampling=" + samplingRate + " polling=" + pollingRate + " -- -- set bridge " + switchName + " sflow=" + agentId
	id, errSetSflowAgentCommand := exec.Command("bash", "-c", setSflowAgentCommand).Output()
	if errSetSflowAgentCommand != nil {
		fmt.Printf("%s", errSetSflowAgentCommand)
	}

	wg.Done() // Need to signal to waitgroup that this goroutine is done
	t := time.Now()
	fmt.Println(t.Format("2006-01-02 15:04:05") + " --- " + "Sflow Agent " + string(id) + " set on " + switchName)
	return string(id)
}

func deleteSflowAgent(switchName string, agentId string, wg *sync.WaitGroup) string {

	deleteSflowAgentCommand := "ovs-vsctl remove bridge " + switchName + " sflow " + agentId
	_, errDeleteSflowAgentCommand := exec.Command("bash", "-c", deleteSflowAgentCommand).Output()
	if errDeleteSflowAgentCommand != nil {
		fmt.Printf("%s", errDeleteSflowAgentCommand)
	}

	wg.Done() // Need to signal to waitgroup that this goroutine is done
	t := time.Now()
	fmt.Println(t.Format("2006-01-02 15:04:05") + " --- " + "Sflow Agent deleted from " + switchName + " switch")
	return "Sflow Agent deleted from " + switchName + " switch"
}

func containerExecCommand(containerCommand ContainerCommand, wg *sync.WaitGroup) string {
	// exec command inside container
	//command := "docker exec -d " + containerCommand.ContainerName + " screen -S " + containerCommand.CommandName + " -d  -m /bin/bash -c '" + containerCommand.Command + " | tee /var/log/" + containerCommand.CommandName + ".log'"
	//command := "docker exec -d " + containerCommand.ContainerName + " nohup "  + containerCommand.Command + " >> /var/log/" + containerCommand.CommandName + ".log &"
	//command := "sudo docker exec -d " + containerCommand.ContainerName + "  bash -c '" +
	//	containerCommand.Command + " > /var/log/" +
	//	containerCommand.CommandName + ".log & echo $!>" +
	//	containerCommand.CommandName+ ".pid' && " + "sudo docker exec " + containerCommand.ContainerName + " tail /" + containerCommand.CommandName + ".pid"
	//_, errExecCommand := exec.Command("bash", "-c", command).Output()
	//if errExecCommand != nil {
	//	fmt.Printf("%s", errExecCommand)
	//}

	commandReadPid := "sudo docker exec " + containerCommand.ContainerName + "  bash -c '" +
		containerCommand.Command + " &> /var/log/" +
		containerCommand.CommandName + ".log & echo $!'"
	cmd := exec.Command("bash", "-c", commandReadPid)
	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput
	err := cmd.Run() // will wait for command to return
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("==> Error: %s\n", err.Error()))
	}
	wg.Done() // Need to signal to waitgroup that this goroutine is done
	t := time.Now()
	fmt.Println(t.Format("2006-01-02 15:04:05") + " --- " + "Command " + containerCommand.CommandName + " is executed Pid is " + string(cmdOutput.Bytes()) + " . Log file available in /var/log/" + containerCommand.CommandName + ".log")
	return "Command " + containerCommand.CommandName + " is executed. Pid -> " + string(cmdOutput.Bytes()) + " [*] Log file is available in /var/log/" + containerCommand.CommandName + ".log"
}

func checkContainerExistByName(container Container, wg *sync.WaitGroup) string {
	commandCheck := "sudo docker ps -a --format '{{.Names}}' | grep -sw " + container.ContainerName
	ifContainerExist, errIfContainerExist := exec.Command("bash", "-c", commandCheck).Output()
	if errIfContainerExist != nil {
		fmt.Printf("%s", errIfContainerExist)
	}
	wg.Done() // Need to signal to waitgroup that this goroutine is done
	t := time.Now()
	if (string(ifContainerExist) != "") {
		fmt.Println(t.Format("2006-01-02 15:04:05") + " --- " + "Countainer " + container.ContainerName + " exist in this host")
		return "true"
	} else {
		fmt.Println(t.Format("2006-01-02 15:04:05") + " --- " + "Countainer " + container.ContainerName + " does not exist in this host")
		return "false"
	}

}

func checkSwitchExistByName(swch Switch, wg *sync.WaitGroup) string {
	commandCheck := "sudo ovs-vsctl show | grep -sw 'Bridge \"" + swch.SwitchName + "\"'"
	ifSwitchExist, errIfSwitchExist := exec.Command("bash", "-c", commandCheck).Output()
	if errIfSwitchExist != nil {
		fmt.Printf("%s", errIfSwitchExist)
	}
	wg.Done() // Need to signal to waitgroup that this goroutine is done
	t := time.Now()
	if (string(ifSwitchExist) != "") {
		fmt.Println(t.Format("2006-01-02 15:04:05") + " --- " + "Switch " + swch.SwitchName + " exist in this host")
		return "true"
	} else {
		fmt.Println(t.Format("2006-01-02 15:04:05") + " --- " + "Switch " + swch.SwitchName + " does not exist in this host")
		return "false"
	}

}

func createDockerContainerWithPortMap(container Container, wg *sync.WaitGroup) string {
	createVnfDocker := "docker run -d  --name " + container.ContainerName + " "
	if (container.Ram != "") {
		createVnfDocker += "--memory=\"" + container.Ram + "m\" "
	}
	if (container.Cpu != "") {
		createVnfDocker += "--cpus=\"" + container.Cpu + "\" "
	}
	if (container.Ports != "") {
		ports := strings.Split(container.Ports, ",")
		for _, port := range ports {
			createVnfDocker += " -p " + port + " "
		}
	}
	createVnfDocker += " " + container.Image + " " + container.InitCommand
	_, errCreateVnfDocker := exec.Command("bash", "-c", createVnfDocker).Output()
	if errCreateVnfDocker != nil {
		fmt.Printf("%s", errCreateVnfDocker)
	}

	fmt.Println(createVnfDocker)
	wg.Done() // Need to signal to waitgroup that this goroutine is done
	t := time.Now()
	fmt.Println(t.Format("2006-01-02 15:04:05") + " --- " + "Container " + container.ContainerName + " Created")
	return "Container " + container.ContainerName + " Created"
}

func updateVNFDocker(container Container, wg *sync.WaitGroup) string {
	updateVnfDocker := "docker update "
	if (container.Ram != "") {
		updateVnfDocker += "--memory=\"" + container.Ram + "m\" "
	}
	if (container.Cpu != "") {
		updateVnfDocker += "--cpus=\"" + container.Cpu + "\" "
	}
	updateVnfDocker += container.ContainerName
	fmt.Println(updateVnfDocker)
	_, errUpdateVnfDocker := exec.Command("bash", "-c", updateVnfDocker).Output()
	if errUpdateVnfDocker != nil {
		fmt.Printf("%s", errUpdateVnfDocker)
	}

	wg.Done() // Need to signal to waitgroup that this goroutine is done
	t := time.Now()
	fmt.Println(t.Format("2006-01-02 15:04:05") + " --- " + "Container " + container.ContainerName + " Created")
	return "Container " + container.ContainerName + " updated"
}

// ************************ Controllers Handler **********************************************************************

func connectToAgetHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func serverStatusHandler(w http.ResponseWriter, r *http.Request) {
	// Execute Command to Create veth pair and connect them to switches
	wg := new(sync.WaitGroup)
	wg.Add(1)
	am := getHostStatus(wg)
	wg.Wait()
	serverStatsJson, err := json.Marshal(am)
	if err != nil {
		fmt.Fprintf(w, "Error: %s", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(serverStatsJson)
	//fmt.Fprintf(w, )
}

func getHostNetworkCardPropertyByNameHandler(w http.ResponseWriter, r *http.Request) {
	networkCard := NetworkCard{} //initialize empty VethPair
	err := json.NewDecoder(r.Body).Decode(&networkCard)
	if err != nil {
		panic(err)
	}
	if len(networkCard.Name) == 0 {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		// Execute Command to Create veth pair and connect them to switches
		wg := new(sync.WaitGroup)
		wg.Add(1)
		am := getHostNetworkCardPropertyByName(networkCard, wg)
		wg.Wait()
		networkcard, err := json.Marshal(am)
		if err != nil {
			fmt.Fprintf(w, "Error: %s", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(networkcard)
	}

}

func createSwitchHandler(w http.ResponseWriter, r *http.Request) {
	swch := Switch{} //initialize empty VethPair
	err := json.NewDecoder(r.Body).Decode(&swch)
	if err != nil {
		panic(err)
	}
	if len(swch.SwitchName) == 0 {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		// Execute Command to Create veth pair and connect them to switches
		wg := new(sync.WaitGroup)
		wg.Add(1)
		am := createSwitch(swch.SwitchName, swch.SwitchControllerIp, swch.SwitchControllerPort, wg)
		wg.Wait()
		fmt.Fprintf(w, am)
	}
}

func deleteSwitchHandler(w http.ResponseWriter, r *http.Request) {
	swch := Switch{} //initialize empty VethPair
	err := json.NewDecoder(r.Body).Decode(&swch)
	if err != nil {
		panic(err)
	}
	if len(swch.SwitchName) == 0 {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		// Execute Command to Create veth pair and connect them to switches
		wg := new(sync.WaitGroup)
		wg.Add(1)
		am := deleteSwitch(swch.SwitchName, wg)
		wg.Wait()
		fmt.Fprintf(w, am)
	}
}

func setSwitchControllerHandler(w http.ResponseWriter, r *http.Request) {
	swch := Switch{} //initialize empty VethPair
	err := json.NewDecoder(r.Body).Decode(&swch)
	if err != nil {
		panic(err)
	}
	if len(swch.SwitchName) == 0 || len(swch.SwitchControllerIp) == 0 || len(swch.SwitchControllerPort) == 0 {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		// Execute Command to Create veth pair and connect them to switches
		wg := new(sync.WaitGroup)
		wg.Add(1)
		am := setSwitchController(swch.SwitchName, swch.SwitchControllerIp, swch.SwitchControllerPort, wg)
		wg.Wait()
		fmt.Fprintf(w, am)
	}
}

func createVethPairHandler(w http.ResponseWriter, r *http.Request) {
	vetPair := VethPair{} //initialize empty VethPair
	err := json.NewDecoder(r.Body).Decode(&vetPair)
	if err != nil {
		panic(err)
	}
	if len(vetPair.SwitchL) == 0 || len(vetPair.SwitchR) == 0 {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		// Execute Command to Create veth pair and connect them to switches
		wg := new(sync.WaitGroup)
		wg.Add(1)
		am := createVethPair(vetPair.SwitchL, vetPair.SwitchR, wg)
		wg.Wait()
		fmt.Fprintf(w, am)
	}
}

func deleteVethPairHandler(w http.ResponseWriter, r *http.Request) {
	vetPair := VethPair{} //initialize empty VethPair
	err := json.NewDecoder(r.Body).Decode(&vetPair)
	if err != nil {
		panic(err)
	}
	if len(vetPair.SwitchL) == 0 || len(vetPair.SwitchR) == 0 {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		// Execute Command to Create veth pair and connect them to switches
		wg := new(sync.WaitGroup)
		wg.Add(1)
		am := deleteVethPair(vetPair.SwitchL, vetPair.SwitchR, wg)
		wg.Wait()
		fmt.Fprintf(w, am)
	}
}

func createVNFDockerHandler(w http.ResponseWriter, r *http.Request) {
	vnfDocker := VnfDocker{} //initialize empty VethPair
	err := json.NewDecoder(r.Body).Decode(&vnfDocker)
	if err != nil {
		panic(err)
	}
	if len(vnfDocker.Name) == 0 || len(vnfDocker.Image) == 0 {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		// Execute Command to Create veth pair and connect them to switches
		wg := new(sync.WaitGroup)
		wg.Add(1)
		am := createVnfDocker(vnfDocker, wg)
		wg.Wait()
		fmt.Fprintf(w, am)
	}
}

func deleteVNFDockerHandler(w http.ResponseWriter, r *http.Request) {
	vnfDocker := VnfDocker{} //initialize empty VethPair
	err := json.NewDecoder(r.Body).Decode(&vnfDocker)
	if err != nil {
		panic(err)
	}
	if len(vnfDocker.Name) == 0 {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		// Execute Command to Create veth pair and connect them to switches
		wg := new(sync.WaitGroup)
		wg.Add(1)
		am := deleteVnfDocker(vnfDocker.Name, wg)
		wg.Wait()
		fmt.Fprintf(w, am)
	}
}

func createOVSDockerPortHandler(w http.ResponseWriter, r *http.Request) {
	oVSDockerPort := OVSDockerPort{} //initialize empty VethPair
	err := json.NewDecoder(r.Body).Decode(&oVSDockerPort)
	if err != nil {
		panic(err)
	}
	if len(oVSDockerPort.VnfName) == 0 || len(oVSDockerPort.SwitchName) == 0 || len(oVSDockerPort.VnfIpAddress) == 0 || len(oVSDockerPort.VnfInterfaceName) == 0 {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		// Execute Command to Create veth pair and connect them to switches
		wg := new(sync.WaitGroup)
		wg.Add(1)
		am := createOVSDockerPort(oVSDockerPort.VnfName, oVSDockerPort.VnfIpAddress, oVSDockerPort.VnfInterfaceName, oVSDockerPort.SwitchName, wg)
		wg.Wait()
		fmt.Fprintf(w, am)
	}
}

func deleteOVSDockerPortHandler(w http.ResponseWriter, r *http.Request) {
	oVSDockerPort := OVSDockerPort{} //initialize empty VethPair
	err := json.NewDecoder(r.Body).Decode(&oVSDockerPort)
	if err != nil {
		panic(err)
	}
	if len(oVSDockerPort.VnfName) == 0 || len(oVSDockerPort.SwitchName) == 0 {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		// Execute Command to Create veth pair and connect them to switches
		wg := new(sync.WaitGroup)
		wg.Add(1)
		am := deleteOVSDockerPort(oVSDockerPort.VnfName, oVSDockerPort.VnfInterfaceName, oVSDockerPort.SwitchName, wg)
		wg.Wait()
		fmt.Fprintf(w, am)
	}
}

func deleteAllOVSDockerPortHandler(w http.ResponseWriter, r *http.Request) {
	oVSDockerPort := OVSDockerPort{} //initialize empty VethPair
	err := json.NewDecoder(r.Body).Decode(&oVSDockerPort)
	if err != nil {
		panic(err)
	}
	if len(oVSDockerPort.VnfName) == 0 || len(oVSDockerPort.SwitchName) == 0 {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		// Execute Command to Create veth pair and connect them to switches
		wg := new(sync.WaitGroup)
		wg.Add(1)
		am := deleteAllOVSDockerPort(oVSDockerPort.VnfName, oVSDockerPort.SwitchName, wg)
		wg.Wait()
		fmt.Fprintf(w, am)
	}
}

func setSflowAgentHandler(w http.ResponseWriter, r *http.Request) {
	sflowAgent := SflowAgent{} //initialize empty VethPair
	err := json.NewDecoder(r.Body).Decode(&sflowAgent)
	if err != nil {
		panic(err)
	}
	if len(sflowAgent.AgentId) == 0 || len(sflowAgent.SwitchName) == 0 || len(sflowAgent.SenderInterface) == 0 || len(sflowAgent.CollectorIp) == 0 || len(sflowAgent.CollectorPort) == 0 || len(sflowAgent.SamplingRate) == 0 || len(sflowAgent.PollingRate) == 0 {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		// Execute Command to Create veth pair and connect them to switches
		wg := new(sync.WaitGroup)
		wg.Add(1)
		am := setSflowAgent(sflowAgent.SwitchName, sflowAgent.AgentId, sflowAgent.SenderInterface, sflowAgent.CollectorIp, sflowAgent.CollectorPort, sflowAgent.SamplingRate, sflowAgent.PollingRate, wg)
		wg.Wait()
		fmt.Fprintf(w, am)
	}
}

func deleteSflowAgentHandler(w http.ResponseWriter, r *http.Request) {
	sflowAgent := SflowAgent{} //initialize empty VethPair
	err := json.NewDecoder(r.Body).Decode(&sflowAgent)
	if err != nil {
		panic(err)
	}
	if len(sflowAgent.AgentId) == 0 || len(sflowAgent.SwitchName) == 0 {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		// Execute Command to Create veth pair and connect them to switches
		wg := new(sync.WaitGroup)
		wg.Add(1)
		am := deleteSflowAgent(sflowAgent.SwitchName, sflowAgent.AgentId, wg)
		wg.Wait()
		fmt.Fprintf(w, am)
	}
}

func getHostStatusHandler(w http.ResponseWriter, r *http.Request) {
	//wg := new(sync.WaitGroup)
	//wg.Add(1)
	//hostStatus := getHostStatus(wg)
	//wg.Wait()
	//fmt.Fprintf(w, hostStatus)

}

func containerExecCommandHandler(w http.ResponseWriter, r *http.Request) {
	containerCommand := ContainerCommand{} //initialize empty VethPair
	err := json.NewDecoder(r.Body).Decode(&containerCommand)
	if err != nil {
		panic(err)
	}
	if len(containerCommand.Command) == 0 || len(containerCommand.ContainerName) == 0 {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		// Execute Command to Create veth pair and connect them to switches
		wg := new(sync.WaitGroup)
		wg.Add(1)
		am := containerExecCommand(containerCommand, wg)
		wg.Wait()
		fmt.Fprintf(w, am)
	}
}

func checkContainerExistByNameHandler(w http.ResponseWriter, r *http.Request) {
	container := Container{} //initialize empty VethPair
	err := json.NewDecoder(r.Body).Decode(&container)
	if err != nil {
		panic(err)
	}
	if len(container.ContainerName) == 0 {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		// Execute Command to Create veth pair and connect them to switches
		wg := new(sync.WaitGroup)
		wg.Add(1)
		am := checkContainerExistByName(container, wg)
		wg.Wait()
		fmt.Fprintf(w, am)
	}
}

func checkSwitchExistByNameHandler(w http.ResponseWriter, r *http.Request) {
	swtch := Switch{} //initialize empty VethPair
	err := json.NewDecoder(r.Body).Decode(&swtch)
	if err != nil {
		panic(err)
	}
	if len(swtch.SwitchName) == 0 {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		// Execute Command to Create veth pair and connect them to switches
		wg := new(sync.WaitGroup)
		wg.Add(1)
		am := checkSwitchExistByName(swtch, wg)
		wg.Wait()
		fmt.Fprintf(w, am)
	}
}

func createDockerContainerWithPortMapHandler(w http.ResponseWriter, r *http.Request) {
	container := Container{} //initialize empty VethPair
	err := json.NewDecoder(r.Body).Decode(&container)
	if err != nil {
		panic(err)
	}
	if len(container.ContainerName) == 0 || len(container.ContainerName) == 0 || len(container.InitCommand) == 0 {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		// Execute Command to Create veth pair and connect them to switches
		wg := new(sync.WaitGroup)
		wg.Add(1)
		am := createDockerContainerWithPortMap(container, wg)
		wg.Wait()
		fmt.Fprintf(w, am)
	}
}

func updateVNFDockerHandler(w http.ResponseWriter, r *http.Request) {
	container := Container{} //initialize empty VethPair
	err := json.NewDecoder(r.Body).Decode(&container)
	if err != nil {
		panic(err)
	}
	if len(container.ContainerName) == 0 {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		// Execute Command to Create veth pair and connect them to switches
		wg := new(sync.WaitGroup)
		wg.Add(1)
		am := updateVNFDocker(container, wg)
		wg.Wait()
		fmt.Fprintf(w, am)
	}
}

func main() {
	fmt.Println(" ")
	fmt.Println("****  XNFV Http Server Agent  ****")
	fmt.Println("****  By AH.GHORAB Fall-2018  ****")
	fmt.Println("****  Version 3.1             ****")
	fmt.Println("----------------------------------")
	fmt.Println("[*] Agent Running at localhost:8000")
	fmt.Println("[*] Valid rest URLs")
	fmt.Println("[#] - /connectToAget")
	fmt.Println("[#] - /serverStatus")
	fmt.Println("[#] - /getHostNetworkCardPropertyByName")
	fmt.Println("[#] - /createSwitch")
	fmt.Println("[#] - /deleteSwitch")
	fmt.Println("[#] - /setSwitchController")
	fmt.Println("[#] - /createVethPair")
	fmt.Println("[#] - /deleteVethPair")
	fmt.Println("[#] - /createVNFDocker")
	fmt.Println("[#] - /updateVNFDocker")
	fmt.Println("[#] - /deleteVNFDocker")
	fmt.Println("[#] - /createDockerContainerWithPortMap")
	fmt.Println("	 - Must fill initial command in the jason request")
	fmt.Println("[#] - /createOVSDockerPort")
	fmt.Println("[#] - /deleteOVSDockerPort")
	fmt.Println("[#] - /deleteALlOVSDockerPort")
	fmt.Println("[#] - /setSflowAgent")
	fmt.Println("[#] - /deleteSflowAgent")
	fmt.Println("[#] - /getHostStatus")
	fmt.Println("[#] - /containerExecCommand")
	fmt.Println("	 - Params: {ContainerName, CommandName, Command}")
	fmt.Println("	 - To stop infinit Commands(like ping) execute -> Kill -9 Command_PID")
	fmt.Println("[#] - /checkContainerExistByName")
	fmt.Println("[#] - /checkSwitchExistByName")
	fmt.Println(" ")
	fmt.Println("------------ Agent Logs ------------")
	fmt.Println(" ")

	// Server status
	http.HandleFunc("/connectToAget", connectToAgetHandler)
	http.HandleFunc("/serverStatus", serverStatusHandler)
	http.HandleFunc("/getHostNetworkCardPropertyByName", getHostNetworkCardPropertyByNameHandler)

	// Create/Delete Switch
	http.HandleFunc("/createSwitch", createSwitchHandler)
	http.HandleFunc("/deleteSwitch", deleteSwitchHandler)
	http.HandleFunc("/setSwitchController", setSwitchControllerHandler)

	// Create/ Delete Veth Pair
	http.HandleFunc("/createVethPair", createVethPairHandler)
	http.HandleFunc("/deleteVethPair", deleteVethPairHandler)

	// Create/Delete VNF Docker
	http.HandleFunc("/createVNFDocker", createVNFDockerHandler)
	http.HandleFunc("/deleteVNFDocker", deleteVNFDockerHandler)
	http.HandleFunc("/updateVNFDocker", updateVNFDockerHandler)
	http.HandleFunc("/createDockerContainerWithPortMap", createDockerContainerWithPortMapHandler)

	// Create/Delete OVS-VNF Docker Ports
	http.HandleFunc("/createOVSDockerPort", createOVSDockerPortHandler)
	http.HandleFunc("/deleteOVSDockerPort", deleteOVSDockerPortHandler)
	http.HandleFunc("/deleteALlOVSDockerPort", deleteAllOVSDockerPortHandler)

	// Set/Delete SFlow Agent
	http.HandleFunc("/setSflowAgent", setSflowAgentHandler)
	http.HandleFunc("/deleteSflowAgent", deleteSflowAgentHandler)

	// Server Statistics
	http.HandleFunc("/getHostStatus", getHostStatusHandler)

	// Execute Command inside Container
	http.HandleFunc("/containerExecCommand", containerExecCommandHandler)

	// Checking Commands
	http.HandleFunc("/checkContainerExistByName", checkContainerExistByNameHandler)
	http.HandleFunc("/checkSwitchExistByName", checkSwitchExistByNameHandler)

	http.ListenAndServe(":8000", nil)

}
