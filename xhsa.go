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
	Name  string `json:"name"`
	Image string `json:"img"`
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
	networkCards []NetworkCard
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
	Image          string `json:"image"`
	ContainerName string `json:"containerName"`
	InitCommand    string `json:"initCommand"`
	Ports          string `json:"ports"`
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

func createVnfDocker(vnfName string, vnfImage string, wg *sync.WaitGroup) string {

	// delete OVS links
	createVnfDocker := "docker run -dit  --name " + vnfName + " --net=none " + vnfImage + " /bin/bash"
	_, errCreateVnfDocker := exec.Command("bash", "-c", createVnfDocker).Output()
	if errCreateVnfDocker != nil {
		fmt.Printf("%s", errCreateVnfDocker)
	}

	wg.Done() // Need to signal to waitgroup that this goroutine is done
	t := time.Now()
	fmt.Println(t.Format("2006-01-02 15:04:05") + " --- " + "VNF " + vnfName + " Created")
	return "VNF " + vnfName + " Created"
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

//func getHostStatus(wg *sync.WaitGroup) HostStatus {
//	hostStatus := HostStatus{}
//
//	// Get Cpu
//	getCpuStatusCommand := "top -bn1 | grep 'Cpu(s)' | sed 's/.*, *\\([0-9.]*\\)%* id.*/\\1/' | awk '{print 100 - $1''}'"
//	cpuStatus, errGetCpuStatusCommand:= exec.Command("bash", "-c", getCpuStatusCommand).Output()
//	if errGetCpuStatusCommand != nil {
//		fmt.Printf("%s", errGetCpuStatusCommand)
//		hostStatus.Cpu = ""
//	} else {
//		hostStatus.Cpu = string(cpuStatus)
//	}
//
//	// Get Ram
//	getRamStatusCommand := "top -bn1 | grep 'Cpu(s)' | sed 's/.*, *\\([0-9.]*\\)%* id.*/\\1/' | awk '{print 100 - $1''}'"
//	ramStatus, errGetRamStatusCommand := exec.Command("bash", "-c", getRamStatusCommand).Output()
//	if errGetRamStatusCommand != nil {
//		fmt.Printf("%s", errGetRamStatusCommand)
//		hostStatus.Ram = ""
//	} else {
//		hostStatus.Ram = string(ramStatus)
//	}
//
//	// Get hhd
//	getHhdStatusCommand := "top -bn1 | grep 'Cpu(s)' | sed 's/.*, *\\([0-9.]*\\)%* id.*/\\1/' | awk '{print 100 - $1''}'"
//	hhdStatus, errGetHhdStatusCommand := exec.Command("bash", "-c", getHhdStatusCommand).Output()
//	if errGetHhdStatusCommand != nil {
//		fmt.Printf("%s", errGetHhdStatusCommand)
//		hostStatus.Cpu = ""
//	} else {
//		hostStatus.Cpu = string(hhdStatus)
//	}
//
//	// Get Ram
//	getCpuUsageCommand := "top -bn1 | grep 'Cpu(s)' | sed 's/.*, *\\([0-9.]*\\)%* id.*/\\1/' | awk '{print 100 - $1''}'"
//	cpuUsage, errGetCpuUsageCommand := exec.Command("bash", "-c", getCpuUsageCommand).Output()
//	if errGetCpuUsageCommand != nil {
//		fmt.Printf("%s", errGetCpuUsageCommand)
//		hostStatus.Cpu = ""
//	} else {
//		hostStatus.Cpu = string(cpuUsage)
//	}
//
//	// Get Ram
//	getCpuStatusCommand := "top -bn1 | grep 'Cpu(s)' | sed 's/.*, *\\([0-9.]*\\)%* id.*/\\1/' | awk '{print 100 - $1''}'"
//	output, errDelOVSLinkL := exec.Command("bash", "-c", getCpuStatusCommand).Output()
//	if errDelOVSLinkL != nil {
//		fmt.Printf("%s", errDelOVSLinkL)
//		hostStatus.Cpu = ""
//	} else {
//		hostStatus.Cpu = string(output)
//	}
//
//	// Get Ram
//	getCpuStatusCommand := "top -bn1 | grep 'Cpu(s)' | sed 's/.*, *\\([0-9.]*\\)%* id.*/\\1/' | awk '{print 100 - $1''}'"
//	output, errDelOVSLinkL := exec.Command("bash", "-c", getCpuStatusCommand).Output()
//	if errDelOVSLinkL != nil {
//		fmt.Printf("%s", errDelOVSLinkL)
//		hostStatus.Cpu = ""
//	} else {
//		hostStatus.Cpu = string(output)
//	}
//
//	// Get Ram
//	getCpuStatusCommand := "top -bn1 | grep 'Cpu(s)' | sed 's/.*, *\\([0-9.]*\\)%* id.*/\\1/' | awk '{print 100 - $1''}'"
//	output, errDelOVSLinkL := exec.Command("bash", "-c", getCpuStatusCommand).Output()
//	if errDelOVSLinkL != nil {
//		fmt.Printf("%s", errDelOVSLinkL)
//		hostStatus.Cpu = ""
//	} else {
//		hostStatus.Cpu = string(output)
//	}
//
//	// Get Ram
//	getCpuStatusCommand := "top -bn1 | grep 'Cpu(s)' | sed 's/.*, *\\([0-9.]*\\)%* id.*/\\1/' | awk '{print 100 - $1''}'"
//	output, errDelOVSLinkL := exec.Command("bash", "-c", getCpuStatusCommand).Output()
//	if errDelOVSLinkL != nil {
//		fmt.Printf("%s", errDelOVSLinkL)
//		hostStatus.Cpu = ""
//	} else {
//		hostStatus.Cpu = string(output)
//	}
//
//	wg.Done() // Need to signal to waitgroup that this goroutine is done
//	t := time.Now()
//	fmt.Println(t.Format("2006-01-02 15:04:05") + " --- " + "Get Host Status")
//	return hostStatus
//}

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

// ************************ Controllers Handler **********************************************************************

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
		am := createVnfDocker(vnfDocker.Name, vnfDocker.Image, wg)
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

func main() {
	fmt.Println(" ")
	fmt.Println("****  XNFV Http Server Agent  ****")
	fmt.Println("****  By AH.GHORAB Fall-2018  ****")
	fmt.Println("****  Version 1.9             ****")
	fmt.Println("----------------------------------")
	fmt.Println("[*] Agent Running at localhost:8000")
	fmt.Println("[*] Valid rest URLs")
	fmt.Println("[#] - /createSwitch")
	fmt.Println("[#] - /deleteSwitch")
	fmt.Println("[#] - /setSwitchController")
	fmt.Println("[#] - /createVethPair")
	fmt.Println("[#] - /deleteVethPair")
	fmt.Println("[#] - /createVNFDocker")
	fmt.Println("[#] - /deleteVNFDocker")
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
