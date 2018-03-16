package util

import "fmt"

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

func KeysToArray(keysMap map[string]bool) []string {
	var keys []string
	for key, _ := range keysMap {
		if !InArray(key, keys) {
			keys = append(keys, key)
		}
	}
	return keys
}

func InArray(value string, array []string) bool {
	for _, k := range array {
		if k == value {
			return true
		}
	}
	return false
}