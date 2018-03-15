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

func KeysToArray(keysMap map[string]bool) (error, []string) {

	keys := []string{}
	for key := range keysMap {
		keys = append(keys, key)
	}

	return nil, keys
}