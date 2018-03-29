package shared

import (
	"../dbStructs"
	"fmt"
	"net/rpc"
	"os"
)

func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

//Handles panic
func HandlePanic() {
	r := recover()
	if r != nil {
		fmt.Println("The following error occured ->", r)
	}
}

//Checks basic errors
func CheckErr(err error) {
	if err != nil {
		panic(err)
	}
}

//Check error without fail stop
func CheckError(err error) error {
	if err != nil {
		fmt.Println("Error ", err.Error())
		return err
	}
	return nil
}

// If error is non-nil, print it out and return it.
func checkError(err error) error {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error ", err.Error())
		return err
	}
	return nil
}

func KeysToArray(keysMap map[string]bool) []string {
	var keys []string
	for key, _ := range keysMap {
		inArray, _ := InArray(key, keys)
		if !inArray {
			keys = append(keys, key)
		}
	}
	return keys
}

func KeysToArray_2(keysMap map[string]*rpc.Client) []string {
	var keys []string
	for key, _ := range keysMap {
		inArray, _ := InArray(key, keys)
		if !inArray {
			keys = append(keys, key)
		}
	}
	return keys
}

func InArray(value string, array []string) (bool, int) {
	for i, k := range array {
		if k == value {
			return true, i
		}
	}
	return false, -1
}

func TableToString(name string, rows map[string]dbStructs.Row) (error, string) {
	tableString := ""
	tableString = tableString + " Table: " + name + "{"

	for key, row := range rows {
		tableString = tableString + "("
		tableString = tableString + "  Key:" + key + " "
		for attribute, value := range row.Data {
			tableString = tableString + attribute + ":" + value + " "
		}
		tableString = tableString + ")"
	}
	tableString = tableString + "}"
	return nil, tableString
}

// Given a table-to-ips mapping, return an ip-to-tables-to-lockOwner mapping
func ReverseMap(peers map[string][]string) (error, map[string]map[string]bool) {
	reverse := make(map[string]map[string]bool)

	for tableName, listOfIps := range peers {
		for _, ip := range listOfIps {

			if _, ok := reverse[ip]; !ok {
				reverse[ip] = make(map[string]bool)
			}

			reverse[ip][tableName] = false

		}
	}

	return nil, reverse
}
