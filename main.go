package main

import "fmt"

func main() {
	//=======================================================
	uname, password, truename := GetConfigData()
	encryptuname := encryptByAES(uname)
	encryptpassword := encryptByAES(password)
	cookie := loginAndGetCookie(encryptuname, encryptpassword)
	//=======================================================
	puid, _ := GetPuid(cookie)
	cookie2, id, name, fullname := GetInfoAndGetCookie2(cookie, puid)
	//=======================================================
	checkCode, uuid := GetCheckCode(cookie)
	formData(checkCode, uuid, cookie2, puid, id, name, fullname, truename)
	fmt.Scanln()
}
