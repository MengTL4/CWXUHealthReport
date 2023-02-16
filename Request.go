package main

import (
	"fmt"
	"github.com/imroc/req/v3"
	"github.com/spf13/viper"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

func GetConfigData() (string, string, string) {
	config := viper.New()
	config.AddConfigPath("./")
	config.SetConfigName("config")
	config.SetConfigType("json")
	if err := config.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("找不到配置文件..")
		} else {
			fmt.Println("配置文件出错..")
		}
	}
	uname := config.GetString("uname")
	password := config.GetString("password")
	truename := config.GetString("truename")
	return uname, password, truename
}

func loginAndGetCookie(uname string, password string) []*http.Cookie {
	client := req.C()
	loginAu, err := client.R().
		SetHeaders(map[string]string{
			"Accept":           "application/json, text/javascript, */*; q=0.01",
			"Origin":           "https://passport2.chaoxing.com",
			"X-Requested-With": "XMLHttpRequest",
			"User-Agent":       "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/109.0",
			"Content-Type":     "application/x-www-form-urlencoded; charset=UTF-8",
			"Referer":          "https://passport2.chaoxing.com/login?fid=&newversion=true&refer=https://i.chaoxing.com",
			"Accept-Encoding":  "gzip, deflate",
			"Accept-Language":  "zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2",
			"Connection":       "close",
		}).
		SetFormData(map[string]string{
			"fid":               "-1",
			"uname":             uname,
			"password":          password,
			"refer":             "https://i.chaoxing.com",
			"t":                 "true",
			"forbidotherlogin":  "0",
			"validate":          "",
			"doubleFactorLogin": "0",
			"independentId":     "0",
		}).
		Post("https://passport2.chaoxing.com/fanyalogin")
	defer loginAu.Body.Close()
	cookies := loginAu.Cookies()
	if err != nil {
		fmt.Println("登陆失败", err)
	}
	return cookies
}

type Response struct {
	Msg struct {
		Puid  int    `json:"puid"`
		Token string `json:"token"`
	} `json:"msg"`
}

func GetPuid(cookies []*http.Cookie) (string, []*http.Cookie) {
	var response Response
	client := req.C()
	get, err := client.R().
		SetHeaders(map[string]string{
			"Accept":           "application/json, text/javascript, */*; q=0.01",
			"Origin":           "https://office.chaoxing.com",
			"X-Requested-With": "XMLHttpRequest",
			"User-Agent":       "Mozilla/5.0 (Linux; Android 5.1.1; NX627J Build/LMY47I; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/52.0.2743.100 Safari/537.36 com.chaoxing.mobile/ChaoXingStudy_3_4.7.4_android_phone_593_53 (@Kalimdor)_9640e4fd20d145319b58cbdb98db86cc",
			"Content-Type":     "application/x-www-form-urlencoded",
			"Referer":          "https://office.chaoxing.com/",
			"Accept-Encoding":  "gzip, deflate",
			"Accept-Language":  "zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2",
			"Connection":       "close",
		}).
		SetCookies(cookies...).
		SetSuccessResult(&response).
		Get("https://noteyd.chaoxing.com/pc/files/getUploadConfig")
	defer get.Body.Close()
	if err != nil {
		fmt.Println("GetPuid失败")
	}
	cookies2 := get.Cookies()
	puid := response.Msg.Puid
	return strconv.Itoa(puid), cookies2
}

type Info struct {
	Data struct {
		UserDept struct {
			Data []struct {
				Id       int    `json:"id"`
				Name     string `json:"name"`
				Fullname string `json:"fullname"`
			} `json:"data"`
		} `json:"userDept"`
	} `json:"data"`
}

func GetInfoAndGetCookie2(cookie []*http.Cookie, puid string) ([]*http.Cookie, string, string, string) {
	var info Info
	client := req.C()
	get, err := client.R().
		SetHeaders(map[string]string{
			"Accept":           "application/json, text/javascript, */*; q=0.01",
			"X-Requested-With": "XMLHttpRequest",
			"User-Agent":       "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/109.0",
			"Referer":          "https://office.chaoxing.com/front/web/apps/forms/fore/apply?id=3449&enc=627c625902a1fd27de56172186a3f903",
			"Accept-Encoding":  "gzip, deflate",
			"Accept-Language":  "zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2",
			"Connection":       "close",
			"Sec-Fetch-Dest":   "empty",
			"Sec-Fetch-Mode":   "cors",
			"Sec-Fetch-Site":   "same-origin",
			"Te":               "trailers",
		}).
		SetCookies(cookie...).
		SetPathParams(map[string]string{
			"uid": puid,
		}).
		SetSuccessResult(&info).
		Get("https://office.chaoxing.com/data/form/find/user/sel/org?uid={uid}")
	defer get.Body.Close()
	if err != nil {
		fmt.Println("GetInfo请求发送失败", err)
	}
	cookie2 := get.Cookies()
	id := info.Data.UserDept.Data[0].Id
	name := info.Data.UserDept.Data[0].Name
	prefullname := info.Data.UserDept.Data[0].Fullname
	var fullname string
	re := regexp.MustCompile(`\\(.+?)\\`)
	match := re.FindStringSubmatch(prefullname)
	if len(match) > 1 {
		fullname = match[1]
	}
	fmt.Println(id, name, fullname)
	return cookie2, strconv.Itoa(id), name, fullname
}

func GetCheckCode(cookie []*http.Cookie) (string, string) {
	client := req.C()
	get, err := client.R().
		SetCookies(cookie...).
		Get("https://office.chaoxing.com/front/web/apps/forms/fore/apply?id=3449&fidEnc=e2869d5221e5bea8&enc=627c625902a1fd27de56172186a3f903")
	if err != nil {
		fmt.Println("GetCheckCode失败", err)
	}
	defer get.Body.Close()
	body, _ := io.ReadAll(get.Body)
	reCheckCode := regexp.MustCompile(`checkCode\s*=\s*'([a-zA-Z0-9]+)'`)
	matchCheckCode := reCheckCode.FindSubmatch(body)
	if matchCheckCode == nil {
		fmt.Println("匹配checkCode失败")
	}
	checkCode := string(matchCheckCode[1])
	fmt.Println("Check code:", checkCode)
	// 使用正则表达式提取 uuid
	reUUID := regexp.MustCompile(`uuid\s*=\s*'([a-zA-Z0-9]+)'`)
	matchUUID := reUUID.FindSubmatch(body)
	if matchUUID == nil {
		fmt.Println("匹配uuid失败")
	}

	uuid := string(matchUUID[1])
	fmt.Println("UUID:", uuid)
	return checkCode, uuid

}
func formData(checkCode string, uuid string, cookies2 []*http.Cookie, puid string, departmentId string, departmentName string, fullname string, truename string) {
	currentDate := time.Now().Format("2006-01-02")
	client := req.C()
	post, err := client.R().
		SetFormData(map[string]string{
			"formId":    "3449",
			"formAppId": "",
			"version":   "9",
			"formData": "[{\"linkInfo\":{\"condFields\":[],\"linkFormType\":\"\",\"linkFormId\":0,\"linkFormValueFieldCompt\":\"\",\"linkFormIdEnc\":\"\",\"linkFormValueFieldId\":0,\"linked\":false},\"compt\":\"contact\",\"loginUserForValue\":true,\"layoutRatio\":1,\"relationValueConfig\":{\"condFieldId\":0,\"type\":0,\"open\":false},\"alias\":\"14\",\"latestValShow\":false,\"optionalScope\":{\"options\":\"\",\"type\":0},\"id\":14,\"fields\":[{\"hasDefaultValue\":true,\"visible\":false,\"editable\":false,\"values\":[{\"puid\":" +
				puid +
				",\"uname\":\"" +
				truename + "\"}],\"name\":\"点击选择人员\",\"verify\":{},\"tip\":{\"imgs\":[],\"text\":\"\"},\"defaultValueStr\":\"[]\",\"label\":\"姓名\",\"sweepCode\":false,\"fieldType\":{\"multiple\":false,\"type\":\"contact\"}}],\"inDetailGroupIndex\":-1,\"fromDetail\":false,\"isShow\":true,\"hasAuthority\":false},{\"linkInfo\":{\"condFields\":[],\"linkFormType\":\"\",\"linkFormId\":0,\"linkFormValueFieldCompt\":\"\",\"linkFormIdEnc\":\"\",\"linkFormValueFieldId\":0,\"linked\":false},\"compt\":\"department\",\"layoutRatio\":1,\"alias\":\"15\",\"id\":15,\"fields\":[{\"hasDefaultValue\":false,\"visible\":false,\"contactId\":14,\"editable\":false,\"values\":[{\"departmentId\":" +
				departmentId + ",\"departmentName\":\"" +
				departmentName + "\",\"serviceId\":-1}],\"name\":\"点击选择部门\",\"verify\":{},\"tip\":{\"imgs\":[],\"text\":\"\"},\"defaultValueStr\":\"[]\",\"label\":\"班级\",\"fieldType\":{\"multiple\":false,\"type\":\"department\"},\"curUserOrg\":false}],\"inDetailGroupIndex\":-1,\"fromDetail\":false,\"isShow\":true,\"hasAuthority\":false},{\"compt\":\"radiobutton\",\"otherAllowed\":false,\"comptCombined\":true,\"layoutRatio\":1,\"alias\":\"13\",\"latestValShow\":true,\"showType\":0,\"id\":13,\"optionColor\":true,\"fields\":[{\"hasDefaultValue\":false,\"visible\":true,\"editable\":true,\"values\":[{\"val\":\"" +
				fullname + "\",\"isOther\":false}],\"options\":[{\"idArr\":[],\"score\":0,\"color\":\"\",\"checked\":false,\"className\":\"color1\",\"title\":\"自动化学院\"},{\"idArr\":[],\"score\":0,\"color\":\"\",\"checked\":false,\"className\":\"color3\",\"title\":\"理学院\"},{\"idArr\":[],\"score\":0,\"color\":\"\",\"checked\":false,\"className\":\"color4\",\"title\":\"轨道交通学院\"},{\"idArr\":[],\"score\":0,\"color\":\"\",\"checked\":true,\"className\":\"color5\",\"title\":\"物联网工程学院\"},{\"idArr\":[],\"score\":0,\"color\":\"\",\"checked\":false,\"className\":\"color6\",\"title\":\"商学院\"},{\"idArr\":[],\"score\":0,\"color\":\"\",\"checked\":false,\"className\":\"color7\",\"title\":\"人文法政学院\"},{\"idArr\":[],\"score\":0,\"color\":\"\",\"checked\":false,\"className\":\"color8\",\"title\":\"传媒与艺术学院\"},{\"idArr\":[],\"score\":0,\"color\":\"\",\"checked\":false,\"className\":\"color9\",\"title\":\"电子信息工程学院\"},{\"idArr\":[],\"score\":0,\"color\":\"\",\"checked\":false,\"className\":\"color10\",\"title\":\"大气与遥感学院\"},{\"idArr\":[],\"score\":0,\"color\":\"\",\"checked\":false,\"className\":\"color1\",\"title\":\"无锡研究生院\"},{\"idArr\":[],\"score\":0,\"color\":\"\",\"checked\":false,\"className\":\"color2\",\"title\":\"环境工程学院\"},{\"idArr\":[],\"score\":0,\"color\":\"#6ee314\",\"checked\":false,\"className\":\"\",\"title\":\"应用技术学院\"}],\"verify\":{\"required\":{}},\"tip\":{\"imgs\":[],\"text\":\"请正确选择自己所属学院\"},\"defaultValueStr\":\"[]\",\"label\":\"学院\",\"fieldType\":{\"type\":\"string\"}}],\"optionScoreShow\":false,\"optionScoreUsed\":false,\"inDetailGroupIndex\":-1,\"fromDetail\":false,\"isShow\":true,\"hasAuthority\":true},{\"linkInfo\":{\"linkFormCondFieldId\":0,\"condFields\":[],\"linkFormType\":\"\",\"linkFormId\":0,\"linkFormValueFieldCompt\":\"\",\"linkFormIdEnc\":\"\",\"linkFormValueFieldId\":0,\"currFormCondFieldId\":0,\"linked\":false},\"compt\":\"numberinput\",\"layoutRatio\":1,\"alias\":\"1\",\"formula\":{\"selIndex\":-1,\"calculateFieldId\":\"0\",\"status\":false},\"latestValShow\":false,\"id\":1,\"fields\":[{\"hasDefaultValue\":false,\"visible\":true,\"percentage\":false,\"values\":[{\"val\":36.6}],\"verify\":{\"minValue\":{\"errMsg\":\"输入的值不能小于\",\"range\":\"\"},\"realNumber\":{\"isInteger\":false,\"precisionLen\":\"\",\"precision\":\"\",\"errMsg\":\"请输入正确的数字类型值\",\"precisionType\":\"0\"},\"maxValue\":{\"errMsg\":\"输入的值不能大于\",\"range\":\"\"},\"required\":{}},\"tip\":{\"imgs\":[],\"text\":\"单位：（C°）\"},\"defaultValueStr\":\"[{\\\"val\\\":\\\"\\\"}]\",\"label\":\"体温上报\",\"placeholderMsg\":\"\",\"fieldType\":{\"type\":\"number\"},\"inputTip\":{\"prefix\":{\"color\":\"#C0C0C3\",\"icon\":\"\"},\"placeholder\":\"\",\"suffix\":{\"color\":\"#C0C0C3\",\"icon\":\"\"}},\"statsable\":false,\"editable\":true}],\"formulaEdit\":{\"formula\":\"\"},\"inDetailGroupIndex\":-1,\"fromDetail\":false,\"isShow\":true,\"hasAuthority\":true},{\"linkInfo\":{\"condFields\":[],\"linkFormType\":\"\",\"linkFormId\":0,\"linkFormValueFieldCompt\":\"\",\"linkFormIdEnc\":\"\",\"linkFormValueFieldId\":0,\"linked\":false},\"compt\":\"editinput\",\"layoutRatio\":1,\"alias\":\"16\",\"formula\":{\"selIndex\":-1,\"calculateFieldId\":\"0\",\"status\":false},\"latestValShow\":false,\"id\":16,\"fields\":[{\"hasDefaultValue\":false,\"visible\":false,\"editable\":false,\"values\":[{\"val\":\"\"}],\"defaultValueStr\":\"[{\\\"val\\\":\\\"\\\"}]\",\"label\":\"中午12点体温上报\",\"codeChangeable\":false,\"associativeInput\":{\"associativeType\":\"\",\"customOption\":{\"options\":[]},\"api\":{\"response\":[{\"jpath\":\"\"}],\"url\":[],\"urlHeaders\":[]},\"bindFormField\":{\"bindFormIdEnc\":\"\",\"bindFieldId\":0,\"bindFieldIdx\":0,\"bindFormType\":\"customForm\",\"bindFormId\":0,\"bindFieldCompt\":\"\"}},\"verify\":{\"charLimit\":{\"maxSize\":42,\"minSize\":33,\"open\":false},\"regularExpress\":{\"errorTip\":\"格式错误!\",\"express\":\"\"},\"unique\":{\"errMsg\":\"此项内容已存在，不允许重复提交\",\"open\":false},\"format\":{\"type\":\"\"}},\"tip\":{\"imgs\":[],\"text\":\"单位：（C°）\"},\"sweepCode\":false,\"fieldType\":{\"type\":\"string\"},\"inputTip\":{\"prefix\":{\"color\":\"#C0C0C3\",\"icon\":\"\"},\"placeholder\":\"\",\"suffix\":{\"color\":\"#C0C0C3\",\"icon\":\"\"}}}],\"formulaEdit\":{\"formula\":\"\"},\"inDetailGroupIndex\":-1,\"fromDetail\":false,\"isShow\":true,\"hasAuthority\":false},{\"compt\":\"radiobutton\",\"otherAllowed\":false,\"comptCombined\":true,\"layoutRatio\":1,\"alias\":\"3\",\"latestValShow\":false,\"showType\":0,\"id\":3,\"optionColor\":false,\"fields\":[{\"hasDefaultValue\":false,\"visible\":true,\"values\":[{\"val\":\"校内\",\"isOther\":false}],\"options\":[{\"idArr\":[],\"score\":0,\"checked\":true,\"title\":\"校内\"},{\"idArr\":[],\"score\":0,\"checked\":false,\"title\":\"校外\"}],\"verify\":{\"required\":{}},\"tip\":{\"imgs\":[],\"text\":\"\"},\"defaultValueStr\":\"[]\",\"label\":\"当前在校内还是校外？\",\"fieldType\":{\"type\":\"string\"},\"editable\":true}],\"optionScoreShow\":false,\"optionScoreUsed\":false,\"inDetailGroupIndex\":-1,\"fromDetail\":false,\"isShow\":true,\"hasAuthority\":true},{\"linkInfo\":{\"condFields\":[],\"linkFormType\":\"customForm\",\"linkFormId\":0,\"linkFormValueFieldCompt\":\"\",\"linkFormIdEnc\":\"\",\"linkFormValueFieldId\":0,\"linked\":false},\"compt\":\"location\",\"locationScope\":{\"linkedInfo\":{},\"mapValue\":[],\"defaultRange\":500,\"select\":false,\"type\":0},\"layoutRatio\":1,\"alias\":\"4\",\"distanceRange\":0,\"id\":4,\"fields\":[{\"visible\":true,\"editable\":true,\"verify\":{\"required\":{}},\"tip\":{\"imgs\":[],\"text\":\"\"},\"label\":\"当前定位\",\"fieldType\":{\"type\":\"point\"},\"values\":[{\"lat\":32.199722,\"lng\":118.708233,\"address\":\"江苏省无锡市锡山区安镇街道锡山大道333号无锡学院\"}]}],\"defaultValueConfig\":0,\"locationValue\":0,\"inDetailGroupIndex\":-1,\"inDetailGroupGeneralId\":-1,\"fromDetail\":false,\"isShow\":true,\"hasAuthority\":true},{\"compt\":\"radiobutton\",\"otherAllowed\":false,\"comptCombined\":true,\"layoutRatio\":1,\"alias\":\"7\",\"latestValShow\":true,\"showType\":0,\"id\":7,\"optionColor\":false,\"fields\":[{\"hasDefaultValue\":false,\"visible\":true,\"values\":[{\"val\":\"否\",\"isOther\":false}],\"options\":[{\"idArr\":[],\"score\":0,\"checked\":false,\"title\":\"是\"},{\"idArr\":[],\"score\":0,\"checked\":true,\"title\":\"否\"}],\"verify\":{\"required\":{}},\"tip\":{\"imgs\":[],\"text\":\"\"},\"defaultValueStr\":\"[]\",\"label\":\"本人目前是否为新冠病毒感染者？\",\"fieldType\":{\"type\":\"string\"},\"editable\":true}],\"optionScoreShow\":false,\"optionScoreUsed\":false,\"inDetailGroupIndex\":-1,\"fromDetail\":false,\"isShow\":true,\"hasAuthority\":true},{\"compt\":\"radiobutton\",\"otherAllowed\":false,\"comptCombined\":true,\"layoutRatio\":1,\"alias\":\"12\",\"latestValShow\":true,\"showType\":0,\"id\":12,\"optionColor\":false,\"fields\":[{\"hasDefaultValue\":false,\"visible\":true,\"editable\":true,\"values\":[{\"val\":\"否\",\"isOther\":false}],\"options\":[{\"idArr\":[],\"score\":0,\"color\":\"\",\"checked\":false,\"className\":\"\",\"title\":\"是\"},{\"idArr\":[],\"score\":0,\"color\":\"\",\"checked\":true,\"className\":\"\",\"title\":\"否\"}],\"verify\":{\"required\":{}},\"tip\":{\"imgs\":[],\"text\":\"\"},\"defaultValueStr\":\"[]\",\"label\":\"本人7天内是否出现过发热、干咳、乏力、嗅味觉减退、鼻塞、流涕、咽痛、结膜炎、肌痛和腹泻等症状？\",\"fieldType\":{\"type\":\"string\"}}],\"optionScoreShow\":false,\"optionScoreUsed\":false,\"inDetailGroupIndex\":-1,\"fromDetail\":false,\"isShow\":true,\"hasAuthority\":true},{\"linkInfo\":{\"linkFormCondFieldId\":0,\"condFields\":[],\"linkFormType\":\"\",\"linkFormId\":0,\"linkFormValueFieldCompt\":\"\",\"linkFormIdEnc\":\"\",\"linkFormValueFieldId\":0,\"currFormCondFieldId\":0,\"linked\":false},\"compt\":\"editinput\",\"layoutRatio\":1,\"alias\":\"10\",\"formula\":{\"selIndex\":-1,\"calculateFieldId\":\"0\",\"status\":false},\"latestValShow\":false,\"id\":10,\"fields\":[{\"hasDefaultValue\":false,\"associativeInput\":{\"associativeType\":\"\",\"customOption\":{\"options\":[]},\"api\":{\"response\":[{\"jpath\":\"\"}],\"url\":[],\"urlHeaders\":[]},\"bindFormField\":{\"bindFormIdEnc\":\"\",\"bindFieldId\":0,\"bindFieldIdx\":0,\"bindFormType\":\"customForm\",\"bindFormId\":0,\"bindFieldCompt\":\"\"}},\"visible\":true,\"values\":[{\"val\":\"\"}],\"verify\":{\"charLimit\":{\"size\":20,\"open\":false},\"regularExpress\":{\"errorTip\":\"格式错误!\",\"express\":\"\"},\"unique\":{\"errMsg\":\"此项内容已存在，不允许重复提交\",\"open\":false},\"format\":{\"type\":\"\"}},\"tip\":{\"imgs\":[],\"text\":\"\"},\"defaultValueStr\":\"[{\\\"val\\\":\\\"\\\"}]\",\"label\":\"其他需要说明的情况\",\"sweepCode\":false,\"fieldType\":{\"type\":\"string\"},\"codeChangeable\":false,\"inputTip\":{\"prefix\":{\"color\":\"#C0C0C3\",\"icon\":\"\"},\"placeholder\":\"\",\"suffix\":{\"color\":\"#C0C0C3\",\"icon\":\"\"}},\"editable\":true}],\"formulaEdit\":{\"formula\":\"\"},\"inDetailGroupIndex\":-1,\"fromDetail\":false,\"isShow\":true,\"hasAuthority\":true},{\"linkInfo\":{\"condFields\":[],\"linkFormType\":\"\",\"linkFormId\":0,\"linkFormValueFieldCompt\":\"\",\"linkFormIdEnc\":\"\",\"linkFormValueFieldId\":0,\"linked\":false},\"compt\":\"dateinput\",\"layoutRatio\":1,\"alias\":\"24\",\"latestValShow\":false,\"id\":24,\"fields\":[{\"hasDefaultValue\":false,\"dayIndex\":0,\"visible\":true,\"editable\":false,\"values\":[{\"val\":\"" +
				currentDate +
				"\"}],\"appoint\":true,\"verify\":{\"validateRange\":{\"beginDate\":\"\",\"dynamicRangeStartValue\":1,\"dynamicRangeEndValue\":1,\"rangeType\":0,\"endDate\":\"\",\"errMsg\":\"请输入合法的日期范围\",\"dynamicRangeStartType\":1,\"dynamicRangeEndType\":3}},\"tip\":{\"imgs\":[],\"text\":\"\"},\"defaultValueStr\":\"[{\\\"val\\\":\\\"\\\"}]\",\"label\":\"提交时间\",\"fieldType\":{\"format\":\"yyyy-MM-dd\",\"type\":\"date\"}}],\"formulaEdit\":{\"formula\":\"\"},\"inDetailGroupIndex\":-1,\"fromDetail\":false,\"isShow\":true,\"hasAuthority\":true}]",
			"ext":             "",
			"t":               "1",
			"enc":             "627c625902a1fd27de56172186a3f903",
			"checkCode":       checkCode,
			"gatherId":        "0",
			"anonymous":       "0",
			"uuid":            uuid,
			"uniqueCondition": "[]",
			"gverify":         "",
		}).
		SetHeaders(map[string]string{
			"Accept":           "application/json, text/javascript, */*; q=0.01",
			"Origin":           "https://office.chaoxing.com",
			"X-Requested-With": "XMLHttpRequest",
			"User-Agent":       "Mozilla/5.0 (Linux; Android 5.1.1; NX627J Build/LMY47I; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/52.0.2743.100 Safari/537.36 com.chaoxing.mobile/ChaoXingStudy_3_4.7.4_android_phone_593_53 (@Kalimdor)_9640e4fd20d145319b58cbdb98db86cc",
			"Content-Type":     "application/x-www-form-urlencoded; charset=UTF-8",
			"Referer":          "https://office.chaoxing.com/front/apps/forms/fore/apply?id=3449&fidEnc=e2869d5221e5bea8&enc=627c625902a1fd27de56172186a3f903",
			"Accept-Encoding":  "gzip, deflate",
			"Accept-Language":  "zh-CN,en-US;q=0.8",
			"Connection":       "close",
		}).
		SetCookies(cookies2...).
		Post("https://office.chaoxing.com/data/apps/forms/fore/user/save?lookuid=198369118")
	defer post.Body.Close()
	if err != nil {
		fmt.Printf("表单提交发生错误%s", err)
	}
	fmt.Println(post)
}
