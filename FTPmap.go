package main

import (
	"fmt"
	"os"
	"os/exec"
	"bufio"
	"net"
	"regexp"
)

func error_check(error_message error){
	if error_message != nil{
		fmt.Println(error_message)
		error_exit("system error occured")
	}
}

func error_exit(message string){
	fmt.Println(message)
	os.Exit(0)
}

func display_help(){
	fmt.Println("FTPmap\n\n" +
				"--help Display application help\n" +
				"-h  Display application help\n" +
				"-t  Specify target host\n" +
				"-s  Specity desitination port, defaults to 21 if port unset.\n" +
				"-u  Specify username (if set password must also be set), defaults to anonymous if unset.\n" + 
				"-p  Specify password\n\n")
	os.Exit(0)
}

func argument_parse() map[string]string {
	var arguments []string = os.Args
	var arguments_count int = len(arguments)

	if arguments_count < 2{
		display_help()
	}

	var parsed_arguments_map = make(map[string]string)

	var target string = "unset" 
	var port string = "unset"
	var username string = "unset"
	var password string = "unset"
	var option string = "unset"

	for argumentIndex, arg := range arguments{
		switch arg{
		case "-h":
			display_help()
		case "--help":
			display_help()
		case "-t":
			target = arguments[argumentIndex + 1]
		case "-s":
			port = arguments[argumentIndex + 1]
		case "-u":
			username = arguments[argumentIndex + 1]
		case "-p":
			password = arguments[argumentIndex + 1]
		case "--bruteforce":
			option = arguments[argumentIndex]
		}
	}

	if username != "unset"{
		if password == "unset"{
			error_exit("Password option must be set.")
		}
	}

	if target == "unset"{
		error_exit("No target has been specified.")
	}

	if port == "unset"{
		port = "21"
	}

	parsed_arguments_map["target"] = target
	parsed_arguments_map["port"] = port
	parsed_arguments_map["username"] = username
	parsed_arguments_map["password"] = password
	parsed_arguments_map["option"] = option

	return parsed_arguments_map
}

func client(option string, remote_connection net.Conn, arguments map[string]string) string{

	//raw command isnt ideal but user will be doing this to their own system... Can always add some regex sanitization...
	ping_output, ping_error := exec.Command("ping","-c 1",arguments["target"]).Output()
	if ping_error != nil{
		error_exit("Unable to reach detination server.")
	}

	//regex to capture ping ttl, we can use this to determine the OS type (sort of)
	pattern := regexp.MustCompile("ttl=(.*?) ")
	ttl_regex := pattern.FindStringSubmatch(string(ping_output))

	var operating_system_ttl string = "Unknown TTL"

	if len(ttl_regex) > 0{
		operating_system_ttl = ttl_regex[len(ttl_regex) - 1]

		if len(operating_system_ttl) == 3{
			if string(operating_system_ttl[0]) == "1"{
				operating_system_ttl = operating_system_ttl + " (Windows)"
			}else if string(operating_system_ttl[0]) == "2"{
				operating_system_ttl = operating_system_ttl + " (Solaris)"
			}
		}else if string(operating_system_ttl[0]) == "6" {
			operating_system_ttl = operating_system_ttl + " (Linux)"
		}else{
			operating_system_ttl = operating_system_ttl + " (Unknown)"
		}
	}
	
	//default to anonymous login is no credentials are set
	if arguments["username"] == "unset"{
		arguments["username"] = "anonymous"
		arguments["password"] = ""
	}

	//attempt login 3 times, packets can get lost afterall
	for i := 0; i < 3; i++{

		fmt.Fprintf(remote_connection, "USER " + arguments["username"] + "\n")
		fmt.Fprintf(remote_connection, "PASS " + arguments["password"] + "\n")
		fmt.Fprintf(remote_connection, "EXIT\n")
		message, _ := bufio.NewReader(remote_connection).ReadString('\n')

		if len(message) > 3{
			if message[:3] == "220"{
				fmt.Print("[-]    " + arguments["target"] + "    " + arguments["port"] + "   " + operating_system_ttl + "   " + message[4:])
			}else if message[:3] == "530"{
				return "[!]    " + arguments["target"] + "    " + arguments["port"] + "   " + arguments["username"] + ":" + arguments["password"]
			}else if message[:3] == "230"{
				return "[+]    " + arguments["target"] + "    " + arguments["port"] + "   " + arguments["username"] + ":" + arguments["password"]
			}
		}
	}
	return "Unknown Status code, exiting...\n\n"
}

func main(){

	arguments := argument_parse()

	remote_connection, remote_connection_error := net.Dial("tcp", arguments["target"] + ":" + arguments["port"])

	error_check(remote_connection_error)

	defer remote_connection.Close()

	var client_response = client(arguments["option"], remote_connection, arguments)

	fmt.Println(client_response)
}
