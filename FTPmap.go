package main

import (
	"fmt"
	"os"
	"bufio"
	"net"
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

	return parsed_arguments_map
}

func authentication_test(remote_connection net.Conn, arguments map[string]string) bool{

	//default to anonymous login is no credentials are set
	if arguments["username"] == "unset"{
		arguments["username"] = "anonymous"
		arguments["password"] = ""
	}

	for i := 0; i < 3; i++{

		fmt.Fprintf(remote_connection, "USER " + arguments["username"] + "\n")
		fmt.Fprintf(remote_connection, "PASS " + arguments["password"] + "\n")
		fmt.Fprintf(remote_connection, "EXIT\n")
		message, _ := bufio.NewReader(remote_connection).ReadString('\n')

		if message[:3] == "530"{
			return false
		}else if message[:3] == "230"{
			return true
		}
	}
	return false
}

func main(){

	arguments := argument_parse()

	remote_connection, remote_connection_error := net.Dial("tcp", arguments["target"] + ":" + arguments["port"])

	error_check(remote_connection_error)

	defer remote_connection.Close()

	if authentication_test(remote_connection, arguments) == false{
		error_exit("Unable to authenticate.")
	}else{
		//work to do...
	}
}
