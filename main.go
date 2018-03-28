package main

import (
	"database/sql"
	"fmt"
	"html/template"
	_ "image"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	//"text/template"
	"github.com/dchest/captcha"
	_ "github.com/go-sql-driver/mysql"
)

const duandian1 = "魑q魅8(噻"
const duandian2 = "簋馥0f娉"

//var User1 = "游客"
var formTemplate = template.Must(template.New("example").Parse(formTemplateSrc))

func addUserToForm(r *http.Request, str string) string {
	r.ParseForm()
	str = strings.Replace(str, "</form>", `<input type="hidden" value="`+r.Form.Get("user")+`" name="user";"/></form>`, -1)
	return str
}
func returnProcess(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fmt.Fprintln(w, `<form action="/process" method=post>
	<input type="hidden" value="`+r.Form.Get("user")+`" name="user";/>
<input type="submit" value="return" name="dowhat";/></form>`)
}

//--------------------------不要脸的函数分割线-----------------------
func showFormHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	d := struct {
		CaptchaId string
	}{
		captcha.NewLen(4),
	}
	if err := formTemplate.Execute(w, &d); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

//跳转界面
func processFormHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Println(err)
	}
	fmt.Println("index页：")
	for k, v := range r.Form {
		fmt.Println("key:", k)
		fmt.Println("val:", strings.Join(v, ""))
	}
	if r.Form.Get("dowhat") == "return" {
		if r.Form.Get("user") == "admin" {
			fmt.Fprintln(w, indexAdmin)
		} else {
			fmt.Fprintln(w, addUserToForm(r, indexHTML))
		}
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if !captcha.VerifyString(r.FormValue("captchaId"), r.FormValue("captchaSolution")) {
		http.Redirect(w, r, "/", http.StatusFound)
	} else {
		stadus := cheakUser(r.FormValue("user"), r.FormValue("usertext"))
		if stadus == 1 {
			fmt.Fprintln(w, addUserToForm(r, indexHTML))
		} else if stadus == 2 {
			fmt.Fprintln(w, indexTr)
		} else if stadus == 3 {
			fmt.Fprintln(w, indexAdmin)
		} else {
			http.Redirect(w, r, "/", http.StatusFound)
		}
		//io.WriteString(w, "Great job, human! You solved the captcha.\n")

	}
	io.WriteString(w, "<br><a href='/'>Try another one</a>")
}

//检查a是否在b的好友名单上是否为好友,1在,0不在
func checkfriends(f string, us string) int {
	db, _ := sql.Open("mysql", "root:qwertyuiop81@tcp(127.0.0.1:3306)/PROJECT_1?parseTime=true")
	var str string
	err := db.QueryRow(`select friends from Users_friends where user="` + us + `"`).Scan(&str)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(str)
	s := strings.Split(str, duandian1)
	fmt.Println(s)
	for _, v := range s {
		if v == f {
			fmt.Println("tamenshihaoyou")
			db.Close()
			return 1
		}
	}
	db.Close()
	return 0
}

//添加好友
func addfriend(f string, us string) {
	db, err := sql.Open("mysql", "root:qwertyuiop81@tcp(127.0.0.1:3306)/PROJECT_1?parseTime=true")
	var str string
	err = db.QueryRow(`select friends from Users_friends where user="` + us + `"`).Scan(&str)
	if err != nil {
		log.Fatal(err)
	}
	s := strings.Split(str, duandian1)
	for _, v := range s {
		if v == f {
			return
		}
	}
	str = str + f + duandian1
	db.Query(`update Users_friends set friends = "` + str + `" where user="` + us + `"`)
	defer db.Close()
}

//删除好友
func delfriend(f string, us string) {
	fmt.Println("开始删除好友")
	db, err := sql.Open("mysql", "root:qwertyuiop81@tcp(127.0.0.1:3306)/PROJECT_1?parseTime=true")
	var str string
	err = db.QueryRow(`select friends from Users_friends where user="` + us + `"`).Scan(&str)
	if err != nil {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		log.Fatal(err)
	}
	str = strings.Replace(str, duandian1+f, "", -1)
	db.Query(`update Users_friends set friends = "` + str + `" where user="` + us + `"`)
	err = db.QueryRow(`select friends from Users_friends where user="` + f + `"`).Scan(&str)
	if err != nil {
		log.SetFlags(log.Lshortfile | log.LstdFlags)
		log.Fatal(err)
	}
	str = strings.Replace(str, duandian1+us, "", -1)
	db.Query(`update Users_friends set friends = "` + str + `" where user="` + f + `"`)
	defer db.Close()
}
func printOldNews(us string, num string) []string {
	ifnew, err1 := strconv.Atoi(num)
	if err1 != nil {
		return nil
	}
	db, err := sql.Open("mysql", "root:qwertyuiop81@tcp(127.0.0.1:3306)/PROJECT_1?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	var str1 string
	err = db.QueryRow(`select chattext from Users_friends where user="` + us + `"`).Scan(&str1)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Query(`update Users_friends set new =0 where user="` + us + `"`)
	if err != nil {
		log.Fatal(err)
	}
	s1 := strings.Split(str1, duandian2)
	s3 := s1[len(s1)-ifnew-1:]
	defer db.Close()
	return s3
}
func friends(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Println(err)
	}
	if f := r.Form.Get("delfr"); f != "" {

		delfriend(f, r.Form.Get("user"))
	}
	if f := r.Form.Get("想添加的好友"); f != "" {
		addfriend(f, r.Form.Get("user"))
	}

	if f := r.Form.Get("oldnews"); f != "" {
		strs := printOldNews(r.Form.Get("user"), r.Form.Get("newsnumber"))
		fmt.Fprintln(w, "<html>")
		for _, v := range strs {
			fmt.Fprintln(w, v+"<br>")
		}
		fmt.Fprintln(w, "</html>")
		return
	}
	fmt.Println("这是Friends的表单：")
	for k, v := range r.Form {
		fmt.Println("key:", k)
		fmt.Println("val:", strings.Join(v, " "))
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	db, err := sql.Open("mysql", "root:qwertyuiop81@tcp(127.0.0.1:3306)/PROJECT_1?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	var ifnew int
	err = db.QueryRow(`select new from Users_friends where user="` + r.Form.Get("user") + `"`).Scan(&ifnew)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(w, "<html><body>")
	rows, _ := db.Query("SELECT * FROM Users_friends")
	db.Close()
	for rows.Next() {
		var addortalk = "添加"
		var tiaozhuan = "myfriends"
		var delfriends string
		var uid int
		var user string
		var friends string
		var chattext string
		var news int
		rows.Scan(&uid, &user, &friends, &chattext, &news)
		if checkfriends(user, r.Form.Get("user")) == 1 && checkfriends(r.Form.Get("user"), user) == 1 {
			addortalk = "聊天"
			delfriends = `<form action="myfriends" method=post style="margin:0px;display:inline;">
            <input type="submit" value="删除" name="删除好友" ;"/> 
            <input type="hidden" value="` + r.FormValue("user") + `" name="user";"/>
            <input type="hidden" value="` + user + `" name="delfr";"/></form>`
			tiaozhuan = "talk"
		} else if checkfriends(user, r.Form.Get("user")) == 0 && checkfriends(r.Form.Get("user"), user) == 1 {
			addortalk = "同意"
		}
		fmt.Fprintln(w, `<form action="`+tiaozhuan+`" method=post style="margin:0px;display:inline;">
            <p>`+user+`<input type="submit" value="`+addortalk+`" name="dowhat" ;"/>
            <input type="hidden" value="`+user+`" name="想添加的好友" ;"/>
            <input type="hidden" value="`+r.FormValue("user")+`" name="user";"/></p></form>`)
		fmt.Fprintln(w, delfriends)
		//checkErr(err)
	}
	if ifnew != 0 {
		fmt.Fprintln(w, `<form action="myfriends" method=post>
            <input type="submit" name="oldnews" value="查看新消息">
            <input type="hidden" value="`+r.FormValue("user")+`" name="user";"/>
            <input type="hidden" name="newsnumber" value="`+strconv.Itoa(ifnew)+`"></form>`)
	}
	fmt.Fprintln(w, `</form><input type="button" name="Submit" value="返回" onclick="javascript:history.go(-1)" /> `)
	fmt.Fprintln(w, "</body></html>")
}
func main() {
	http.HandleFunc("/", showFormHandler)
	http.HandleFunc("/process", processFormHandler)
	http.HandleFunc("/uploadgra", uploadgra)
	http.HandleFunc("/uploadtxt", uploadtxt)
	http.HandleFunc("/loadtxt", loadtxt)
	http.HandleFunc("/loadgra", loadwhichgra)
	http.HandleFunc("/gra", loadgra)
	http.HandleFunc("/myfriends", friends)
	http.HandleFunc("/talk", talkfriends)
	http.HandleFunc("/textdir", textdir)
	http.HandleFunc("/textdir/doc", textdoc)
	http.HandleFunc("/searchfiles", searchfiles)
	http.HandleFunc("/readannouncement", readAnnouncement)
	//管理员权限网站
	http.HandleFunc("/delgra", delgra)
	http.HandleFunc("/deltext", deltext)
	http.HandleFunc("/deluser", deluser)
	http.HandleFunc("/delcommend", delcommend)
	http.HandleFunc("/addannouncement", addAnnouncement)
	http.Handle("/captcha/", captcha.Server(captcha.StdWidth, captcha.StdHeight))
	fmt.Println("Server is at localhost:8080")
	if err := http.ListenAndServe("localhost:8080", nil); err != nil {
		log.Fatal(err)
	}
}

//管理员权限区-------------------------------------
func delgra(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fmt.Println("删图：")
	for k, v := range r.Form {
		fmt.Println("key:", k)
		fmt.Println("val:", strings.Join(v, " "))
	}
	if r.Method == "POST" && r.Form.Get("user") == "admin" {
		if vi := r.Form.Get("删"); vi != "" {
			os.Remove("./uploadimg/" + vi)
		}
		files, _ := ioutil.ReadDir("./uploadimg")
		fmt.Fprintln(w, `<html>`)
		for _, v := range files {
			fmt.Fprintln(w, v.Name()+`<form action="/delgra" method=post><input type="submit" name="删除" value="del">
		<input type="hidden" name="删" value="`+v.Name()+`";/>
		<input type="hidden" name="user" value="`+r.Form.Get("user")+`";><br></form>`)
		}
		returnProcess(w, r)
		fmt.Fprintln(w, `/<html>`)
	} else {
		http.Redirect(w, r, "/", http.StatusFound)
	}
}
func deltext(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fmt.Println("删文：")
	for k, v := range r.Form {
		fmt.Println("key:", k)
		fmt.Println("val:", strings.Join(v, " "))
	}
	if r.Method == "POST" && r.Form.Get("user") == "admin" {
		db, err := sql.Open("mysql", "root:qwertyuiop81@tcp(127.0.0.1:3306)/PROJECT_1?parseTime=true")
		if err != nil {
			log.Fatal(err)
		}
		if vi := r.Form.Get("删"); vi != "" {
			_, err := db.Exec(`DELETE  FROM Users_text WHERE artical_name ="` + vi + `" `)
			_, err = db.Exec(`DELETE  FROM Users_commend WHERE artical_name ="` + vi + `" `)
			if err != nil {
				return
			}
		}
		rows, _ := db.Query("SELECT artical_name FROM Users_text")
		var k []string
		for rows.Next() {
			var atcN string
			rows.Scan(&atcN)
			//checkErr(err)
			k = append(k, atcN)
		}
		fmt.Fprintln(w, "<html><head></head><body>")
		for i := 0; i < len(k); i++ {
			fmt.Fprintln(w, `<form action="/deltext" method=post><p>`+k[i]+`<input type="submit"  name="删除" value="删" ;"/>
			<input type="hidden" name="删" value="`+k[i]+`";>
			<input type="hidden" value="`+r.FormValue("user")+`" name="user";"/></form> `)
		}
		returnProcess(w, r)
		fmt.Fprintln(w, "</body></html>")
		defer db.Close()
	} else {
		http.Redirect(w, r, "/", http.StatusFound)
	}
}
func deluser(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fmt.Println("删除用户：")
	for k, v := range r.Form {
		fmt.Println("key:", k)
		fmt.Println("val:", strings.Join(v, " "))
	}
	if r.Method == "POST" && r.Form.Get("user") == "admin" {
		db, err := sql.Open("mysql", "root:qwertyuiop81@tcp(127.0.0.1:3306)/PROJECT_1?parseTime=true")
		if err != nil {
			log.Fatal(err)
		}
		if vi := r.Form.Get("删"); vi != "" {
			_, err := db.Exec(`DELETE  FROM Users_normal WHERE username ="` + vi + `" `)
			if err != nil {
				return
			}
		}
		rows, _ := db.Query("SELECT username FROM Users_normal")
		var k []string
		for rows.Next() {
			var user string
			rows.Scan(&user)
			//checkErr(err)
			k = append(k, user)
		}
		fmt.Fprintln(w, "<html><head></head><body>")
		for i := 0; i < len(k); i++ {
			fmt.Fprintln(w, `<form action="/deluser" method=post><p>`+k[i]+`<input type="submit"  name="删除" value="删" ;"/>
			<input type="hidden" name="删" value="`+k[i]+`";>
			<input type="hidden" value="`+r.FormValue("user")+`" name="user";"/></form> `)
		}
		returnProcess(w, r)
		fmt.Fprintln(w, "</body></html>")
		defer db.Close()
	} else {
		http.Redirect(w, r, "/", http.StatusFound)
	}
}
func delcommend(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fmt.Println("删除评论：")
	for k, v := range r.Form {
		fmt.Println("key:", k)
		fmt.Println("val:", strings.Join(v, " "))
	}
	if r.Method == "POST" && r.Form.Get("user") == "admin" {
		var commendAll []string
		var commendAllID []int
		db, err := sql.Open("mysql", "root:qwertyuiop81@tcp(127.0.0.1:3306)/PROJECT_1?parseTime=true")
		if err != nil {
			log.Fatal(err)
		}
		if vi := r.Form.Get("删"); vi != "" {
			temp, _ := strconv.Atoi(vi)
			err := delDircommend(temp)
			if err != nil {
				fmt.Fprintln(w, "<html><head></head><body>")
				returnProcess(w, r)
				return
			}
		}
		rows, _ := db.Query("SELECT id,commend  FROM Users_commend")
		for rows.Next() {
			var id int
			var delcommends string
			rows.Scan(&id, &delcommends)
			//checkErr(err)
			commendAllID = append(commendAllID, id)
			commendAll = append(commendAll, delcommends)
		}
		fmt.Fprintln(w, "<html><head></head><body>")
		for i := 0; i < len(commendAll); i++ {
			fmt.Fprintln(w, `<form action="/delcommend" method=post><p>`+commendAll[i]+`<input type="submit"  name="删除" value="删" ;"/>
			<input type="hidden" name="删" value="`+strconv.Itoa(commendAllID[i])+`";>
			<input type="hidden" value="`+r.FormValue("user")+`" name="user";"/></form> `)
		}
		returnProcess(w, r)
		fmt.Fprintln(w, "</body></html>")
		defer db.Close()
	} else {
		http.Redirect(w, r, "/", http.StatusFound)
	}
}
func delDircommend(id int) error {
	db, err := sql.Open("mysql", "root:qwertyuiop81@tcp(127.0.0.1:3306)/PROJECT_1?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	rows, _ := db.Query("SELECT id  FROM Users_commend WHERE parentid=" + strconv.Itoa(id) + "")
	//delDircommend(-1,id)
	for rows.Next() {
		var sonid int
		err := rows.Scan(&sonid)
		//checkErr(err)
		if err == nil {
			delDircommend(sonid)
		}

	}
	_, err = db.Exec(`DELETE  FROM Users_commend WHERE id =` + strconv.Itoa(id) + ``)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	return nil
}
func addAnnouceToMYSQL(title string, str string) {
	db, err := sql.Open("mysql", "root:qwertyuiop81@tcp(127.0.0.1:3306)/PROJECT_1?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(`insert into Users_announce (title,announcement,submission_date) VALUES  ("` + title + `","` + str + `",NOW());`)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
}
func addAnnouncement(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if r.Method == "POST" && r.Form.Get("user") == "admin" {
		if r.Form.Get("公告标题") != "" && r.Form.Get("公告内容") != "" {
			addAnnouceToMYSQL(r.Form.Get("公告标题"), r.Form.Get("公告内容"))
		}
		fmt.Fprintln(w, addUserToForm(r, addannouncement))
		returnProcess(w, r)
	} else {
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

//管理员权限区-------------------------------------
func readAnnouncement(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Println(err)
	}
	fmt.Println("读公告：")
	for k, v := range r.Form {
		fmt.Println("key:", k)
		fmt.Println("val:", strings.Join(v, ""))
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	db, err := sql.Open("mysql", "root:qwertyuiop81@tcp(127.0.0.1:3306)/PROJECT_1?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	rows, _ := db.Query("SELECT title,announcement,submission_date  FROM Users_announce")

	fmt.Fprintln(w, "<html><head></head><body>")
	for rows.Next() {
		var title string
		var announceMent string
		var date string
		rows.Scan(&title, &announceMent, &date)
		fmt.Fprintln(w, `<h2>`+title+`</h2>`+date+`<br>`)
		fmt.Fprintln(w, announceMent+"<hr><hr>")
	}
	returnProcess(w, r)
	fmt.Fprintln(w, "</body></html>")
	defer db.Close()
}
func foundtxtText(str string, way string) []string {
	db, err := sql.Open("mysql", "root:qwertyuiop81@tcp(127.0.0.1:3306)/PROJECT_1?parseTime=true")
	var getwhat string
	if err != nil {
		log.Fatal(err)
	}
	switch way {
	case "1":
		getwhat = "artical_name"
	case "2":
		getwhat = "artical"
	case "3":
		getwhat = "submission_date"
	case "4":
		getwhat = "author"
	}
	rows, err := db.Query(`SELECT id,` + getwhat + ` FROM Users_text`)
	if err != nil {
		log.Fatal(err)
	}
	var match []int
	for rows.Next() {
		var id int
		var found string
		err := rows.Scan(&id, &found)
		if err != nil {
			log.Fatal(err)
		}
		//这个Rabin-Karp算法很好的zzz
		if strings.Contains(found, str) {
			match = append(match, id)
		}
	}
	var matchname []string
	for _, v := range match {
		var str1 string
		err := db.QueryRow(`SELECT artical_name FROM Users_text WHERE id=` + strconv.Itoa(v) + ``).Scan(&str1)
		if err != nil {
			continue
		}
		matchname = append(matchname, str1)
	}

	return matchname
}
func searchfiles(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	var found []string
	if r.Form.Get("查找内容") != "" {
		found = foundtxtText(r.Form.Get("查找内容"), r.Form.Get("查找方式"))
		fmt.Println(found)
	}
	fmt.Println("查找文件：：")
	for k, v := range r.Form {
		fmt.Println("key:", k)
		fmt.Println("val:", strings.Join(v, " "))
	}
	fmt.Fprintln(w, addUserToForm(r, findTextDirHTML))
	for _, v := range found {
		fmt.Fprintln(w, v+"<br>")
	}
}

//多级评论系统递归算法实现
var commend []string

func loadcommend(parentid string, atcN string) []string {
	db, err := sql.Open("mysql", "root:qwertyuiop81@tcp(127.0.0.1:3306)/PROJECT_1?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	rows, err := db.Query(`SELECT id,commend FROM Users_commend where artical_name="` + atcN + `" AND parentid="` + parentid + `"`)
	if err != nil {
		log.Fatal(err)
	}
	db.Close()
	for rows.Next() {
		var id int
		var com string
		rows.Scan(&id, &com)
		if parentid != "0" {
			commend = append(commend, "&nbsp&nbsp&nbsp&nbsp"+strconv.Itoa(id)+" "+com)
		} else {
			commend = append(commend, strconv.Itoa(id)+" "+com)
		}

		loadcommend(strconv.Itoa(id), atcN)
	}
	return commend
}
func commentothers(parentid string, us string, comments string, text string) {
	db, err := sql.Open("mysql", "root:qwertyuiop81@tcp(127.0.0.1:3306)/PROJECT_1?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	var str string
	if parentid != "0" {
		err = db.QueryRow(`select person from Users_commend where artical_name="` + text + `" AND id="` + parentid + `"`).Scan(&str)
		if err != nil {
			log.Println(err)
			return
		}
	}
	_, err = db.Exec(`insert into Users_commend (parentid,artical_name,commend,person,submission_date) VALUES  (` + parentid + `,"` + text + `","` + us + ` 回复 ` + str + `: ` + comments + `","` + us + `",NOW());`)
	if err != nil {
		log.Fatal(err)
	}
	db.Close()
}
func textdoc(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if r.Form.Get("回复内容") != "" {
		if l := r.Form.Get("楼层"); l != "" {

			commentothers(l, r.FormValue("user"), r.Form.Get("回复内容"), r.Form.Get("查看"))

		} else {
			commentothers("0", r.FormValue("user"), r.Form.Get("回复内容"), r.Form.Get("查看"))
		}
	}
	fmt.Println("文档：")
	for k, v := range r.Form {
		fmt.Println("key:", k)
		fmt.Println("val:", strings.Join(v, " "))
	}
	db, err := sql.Open("mysql", "root:qwertyuiop81@tcp(127.0.0.1:3306)/PROJECT_1?parseTime=true")
	if r.Form.Get("查看") != "" {
		var str string
		err = db.QueryRow(`select artical from Users_text where artical_name="` + r.Form.Get("查看") + `"`).Scan(&str)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprintln(w, "<html><head></head><body>")
		s := strings.Split(str, "\n")
		for _, v := range s {
			io.WriteString(w, v)
			io.WriteString(w, "<br>")
		}
		fmt.Fprintln(w, "<hr><hr><h3>评论：</h3><br>")
		strs := loadcommend("0", r.Form.Get("查看"))
		for _, v := range strs {
			fmt.Fprintln(w, v+"<br>")
		}
		//WOW有问题
		commend = commend[:0]
		fmt.Fprintln(w, `回复：<form action="/textdir/doc" method=post>`)
		//限制游客不能回复别人
		if r.Form.Get("user") != "游客" {
			fmt.Fprintln(w, `<input type="text" name="楼层" size="10px";/>`)
		}
		fmt.Fprintln(w, `<input type="text" name="回复内容"/><input type="hidden" value="`+r.FormValue("user")+`" name="user";/>
            <input type="hidden" value="`+r.FormValue("查看")+`" name="查看";/><input type="submit" name="发表" ;/></form>`)

		fmt.Fprintln(w, `<input type="button" name="Submit" value="返回" onclick="javascript:history.go(-1)" /> `)
		fmt.Fprintln(w, "</body></html>")
		defer db.Close()
	}
}
func textdir(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fmt.Println("目录：")
	for k, v := range r.Form {
		fmt.Println("key:", k)
		fmt.Println("val:", strings.Join(v, " "))
	}
	db, err := sql.Open("mysql", "root:qwertyuiop81@tcp(127.0.0.1:3306)/PROJECT_1?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	rows, _ := db.Query("SELECT artical_name FROM Users_text")
	var k []string
	for rows.Next() {
		var atcN string
		rows.Scan(&atcN)
		//checkErr(err)
		k = append(k, atcN)
	}
	fmt.Fprintln(w, "<html><head></head><body>")
	for i := 0; i < len(k); i++ {
		fmt.Fprintln(w, `<form action="/textdir/doc" method=post><p>`+k[i]+`<input type="submit"  name="查看" value="`+k[i]+`" ;"/>
			<input type="hidden" value="`+r.FormValue("user")+`" name="user";"/></form> `)
	}
	fmt.Fprintln(w, `<input type="button" name="Submit" value="返回" onclick="javascript:history.go(-1)" /> `)
	fmt.Fprintln(w, "</body></html>")
	defer db.Close()
}
func addchattext(fri string, us string, chat string) {
	chat = us + ":" + fri + chat
	db, err := sql.Open("mysql", "root:qwertyuiop81@tcp(127.0.0.1:3306)/PROJECT_1?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	var str string
	err = db.QueryRow(`select chattext from Users_friends where user="` + us + `"`).Scan(&str)
	if err != nil {
		log.Fatal(err)
	}
	str = str + chat + duandian2
	_, err = db.Query(`update Users_friends set chattext = "` + str + `" where user="` + us + `"`)
	if err != nil {
		log.Fatal(err)
	}
	err = db.QueryRow(`select chattext from Users_friends where user="` + fri + `"`).Scan(&str)
	if err != nil {
		log.Fatal(err)
	}
	str = str + chat + duandian2
	_, err = db.Query(`update Users_friends set chattext = "` + str + `" where user="` + fri + `"`)
	if err != nil {
		log.Fatal(err)
	}
	var newnum int
	err = db.QueryRow(`select new from Users_friends where user="` + fri + `"`).Scan(&newnum)
	if err != nil {
		log.Fatal(err)
	}
	newnum++
	_, err = db.Query(`update Users_friends set new=` + strconv.Itoa(newnum) + ` where user="` + fri + `"`)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
}
func talkfriends(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Println(err)
	}
	if k := r.Form.Get("news"); k != "" {
		fmt.Println("向数据库写入聊天记录")
		fmt.Println(r.Form.Get("想添加的好友"))
		fmt.Println(r.Form.Get("user"))
		fmt.Println(r.Form.Get("news"))
		addchattext(r.Form.Get("想添加的好友"), r.Form.Get("user"), r.Form.Get("news"))
	}
	fmt.Println("这是两人聊天的表单：")
	for k, v := range r.Form {
		fmt.Println("key:", k)
		fmt.Println("val:", strings.Join(v, " "))
	}
	db, _ := sql.Open("mysql", "root:qwertyuiop81@tcp(127.0.0.1:3306)/PROJECT_1?parseTime=true")

	var str1 string
	var str2 string
	err := db.QueryRow(`select chattext from Users_friends where user="` + r.Form.Get("user") + `"`).Scan(&str1)
	if err != nil {
		log.Fatal(err)
	}
	err = db.QueryRow(`select chattext from Users_friends where user="` + r.Form.Get("想添加的好友") + `"`).Scan(&str2)
	if err != nil {
		log.Fatal(err)
	}
	//增加消息提醒系统
	s1 := strings.Split(str1, duandian2)
	s2 := strings.Split(str2, duandian2)
	fmt.Println(s1)
	fmt.Println(s2)
	var str []string
	//GOODPOINT
	for i := 0; i < len(s1); i++ {
		for j := 0; j < len(s2); j++ {
			if s1[i] == s2[j] && s1[i] != "" {
				str = append(str, s1[i])
				s1 = append(s1[:i], s1[i+1:]...)
			}
		}
	}
	fmt.Fprintln(w, "<html><head></head><body><p>")
	for _, v := range str {
		fmt.Fprintln(w, v+"<br>")
	}
	fmt.Fprintln(w, `</p><form action="talk" method=post><input type="textarea"  name="news"  ;"/>
        <input type="hidden" value="`+r.FormValue("想添加的好友")+`" name="想添加的好友" ;"/>
        <input type="hidden" value="`+r.FormValue("user")+`" name="user";"/>
        <input type="submit" value="发送">`)
	fmt.Fprintln(w, `</form><input type="button" name="Submit" value="返回" onclick="javascript:history.go(-1)" /> `)
	fmt.Fprintln(w, "</body></html>")
	db.Close()
}

//1为新用户或登录 2为游客
func cheakUser(un string, uw string) int {
	if un == "admin" && uw == "qwertyuiop81" {
		return 3
	}
	db, err := sql.Open("mysql", "root:qwertyuiop81@tcp(127.0.0.1:3306)/PROJECT_1?parseTime=true")
	fmt.Println(un, uw)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS Users_normal(
        id INT UNSIGNED AUTO_INCREMENT,
       username VARCHAR(100) NOT NULL,
        userpw VARCHAR(40) NOT NULL,
        submission_date DATE,
        PRIMARY KEY (id )
     )ENGINE=InnoDB DEFAULT CHARSET=utf8;`)
	if err != nil {
		log.Fatalln(err)
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS Users_friends(
        id INT UNSIGNED AUTO_INCREMENT,
        user VARCHAR(100) NOT NULL,
       friends TEXT ,
        chattext TEXT ,
        PRIMARY KEY (id )
     )ENGINE=InnoDB DEFAULT CHARSET=utf8;`)
	if err != nil {
		log.Fatalln(err)
	}
	rows, _ := db.Query("SELECT * FROM Users_normal")
	for rows.Next() {
		var uid int
		var username string
		var userpw string
		var data string
		rows.Scan(&uid, &username, &userpw, &data)
		if username == un && userpw != uw {
			return 3
		} else if username == un && userpw == uw {
			return 1
		}
		//checkErr(err)
		fmt.Println(uid)
		fmt.Println(username)
		fmt.Println(userpw)
		fmt.Println(data)
	}
	if un != "" && uw != "" {
		x := `INSERT INTO Users_normal
        (username,userpw,submission_date)
        VALUES
        ("` + un + `","` + uw + `",NOW());`
		y := `INSERT INTO Users_friends
        (user,friends,chattext,new)
        VALUES
        ("` + un + `","` + duandian1 + `","` + duandian2 + `",0);`
		//x := "INSERT INTO Users_normal(username,userpw,submission_date) VALUES (" + un + "," + uw + ",NOW())"
		_, err := db.Exec(x)
		if err != nil {
			log.Fatalln(err)
		}
		_, err = db.Exec(y)
		if err != nil {
			log.Fatalln(err)
		}
		/*_, err := db.Exec(x)
		if err != nil {
			log.Fatalln(err)
		}*/
		db.Close()
		return 1
	} else {
		db.Close()
		return 2
	}
}

func uploadtxt(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, `<html lang="en">
        <head>
            <meta charset="UTF-8">
            <title>Document</title>
        </head>
        <body>
            <form action="/uploadtxt" method="post" enctype="multipart/form-data">
                文件：<input type="file" name="file" value="">
                <input type="hidden" value="`+r.FormValue("user")+`" name="user";"/>
                <input type="submit" value="提交">
            </form>
        </body>
        </html>`)
	fmt.Fprintln(w, `<input type="button" name="Submit" value="返回" onclick="javascript:history.go(-1)" /></html> `)
	if r.Method == "POST" {
		r.ParseMultipartForm(32 << 20)
		r.ParseForm()
		for k, v := range r.Form {
			fmt.Println("key:", k)
			fmt.Println("val:", strings.Join(v, ""))
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()
		db, err := sql.Open("mysql", "root:qwertyuiop81@tcp(127.0.0.1:3306)/PROJECT_1?parseTime=true")
		if err != nil {
			log.Fatal(err)
		}
		_, err = db.Exec(`CREATE TABLE IF NOT EXISTS Users_text(
        id INT UNSIGNED AUTO_INCREMENT,
       artical TEXT NOT NULL,
       artical_name VARCHAR(100) NOT NULL,
        submission_date DATE,
        PRIMARY KEY (id )
     )ENGINE=InnoDB DEFAULT CHARSET=utf8;`)
		if err != nil {
			log.Fatalln(err)
		}
		buffer1 := make([]byte, 32384)
		file.Read(buffer1)
		k := string(buffer1)
		fmt.Println(k)
		x := `INSERT INTO Users_text
        (artical,artical_name,submission_date,author)
        VALUES
        ("` + k + `","` + header.Filename + `",NOW(),"ceshi5");`
		_, err = db.Exec(x)
		if err != nil {
			return
		}
		db.Close()
	}
}
func loadtxt(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Println(err)
	}
	for k, v := range r.Form {
		fmt.Println("key:", k)
		fmt.Println("val:", strings.Join(v, ""))
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	db, err := sql.Open("mysql", "root:qwertyuiop81@tcp(127.0.0.1:3306)/PROJECT_1?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	rows, _ := db.Query("SELECT artical FROM Users_text")
	var k []string
	for rows.Next() {
		var atc string
		rows.Scan(&atc)
		//checkErr(err)
		k = append(k, atc)
	}
	//fmt.Println(k[1])
	//	k[1] = strings.Replace(k[1], "\n", "\r\n", -1)
	//k[1] = strings.Replace(k[1], "\n", "\r\n", -1)
	//fmt.Fprintln(w, k[1])
	fmt.Fprintln(w, "<html><head></head><body>")
	for i := 0; i < len(k); i++ {
		str := strings.Split(k[i], "\n")
		for _, v := range str {
			io.WriteString(w, v)
			io.WriteString(w, "<br>")
		}
		io.WriteString(w, "<hr><hr>")
	}
	fmt.Fprintln(w, `<input type="button" name="Submit" value="返回" onclick="javascript:history.go(-1)" /> `)
	fmt.Fprintln(w, "</body></html>")
	db.Close()
}

func uploadgra(w http.ResponseWriter, r *http.Request) {
	//判断请求方式
	fmt.Fprintln(w, uploadgraph)
	fmt.Fprintln(w, `<input type="button" name="Submit" value="返回" onclick="javascript:history.go(-1)" /></html> `)
	if r.Method == "POST" {
		fmt.Println("一一一一")
		//设置内存大小
		r.ParseMultipartForm(32 << 30)
		//获取上传的第一个文件
		file, header, err := r.FormFile("img")
		if err != nil {
			fmt.Println("一一a一一")
			log.Println(err)
			return
		}
		//创建上传目录
		os.Mkdir("./uploadimg", os.ModePerm)
		//创建上传文件
		cur, err := os.Create("./uploadimg/" + header.Filename)
		if err != nil {
			fmt.Println("一一b一一")
			log.Fatal(err)
		}
		//把上传文件数据拷贝到我们新建的文件
		io.Copy(cur, file)
		defer file.Close()
		defer cur.Close()
	}

}

var seeWhichPicture int

func loadwhichgra(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, `<html><form action="/gra" method=post>
        <input type="text" name="id">
        <input type="submit" value="查看" name="find" ;"/>
        <input type="hidden" value="`+r.FormValue("user")+`" name="user";"/></form><br><br>`)
}
func loadgra(w http.ResponseWriter, r *http.Request) {
	//判断请求方式
	//好像、、可以用writeimage解决？？？
	r.ParseForm()
	id := r.Form.Get("id")
	p, err := strconv.Atoi(id)
	if err != nil || p < 1 {
		return
	}
	files, _ := ioutil.ReadDir("./uploadimg")
	if p >= len(files) {
		p = p%len(files) + 1
	}
	fi := files[p]
	sli := strings.Split(fi.Name(), ".")
	if sli[len(sli)-1] == "png" {
		k := "./uploadimg/" + fi.Name()
		fmt.Println(k)
		fil, err := os.Open(k)
		if err != nil {
			fmt.Println("question1")
			log.Fatal(err)
		}
		file, err := png.Decode(fil)
		if err != nil {
			fmt.Println("question2")
			log.Fatal(err)
		}
		png.Encode(w, file)
		defer fil.Close()
	} else if sli[len(sli)-1] == "jpeg" {
		k := "./uploadimg/" + fi.Name()
		fmt.Println(k)
		fil, err := os.Open(k)
		if err != nil {
			fmt.Println("question1")
			log.Fatal(err)
		}
		file, err := jpeg.Decode(fil)
		if err != nil {
			fmt.Println("question2")
			log.Fatal(err)
		}
		jpeg.Encode(w, file, nil)
		defer fil.Close()
	}
}

const formTemplateSrc = `<!doctype html>
<head><title>Captcha Example</title></head>
<body style="text-align:center;">
<script>
function setSrcQuery(e, q) {
    var src  = e.src;
    var p = src.indexOf('?');
    if (p >= 0) {
        src = src.substr(0, p);
    }
    e.src = src + "?" + q
}
function playAudio() {
    var le = document.getElementById("lang");
    var lang = le.options[le.selectedIndex].value;
    var e = document.getElementById('audio')
    setSrcQuery(e, "lang=" + lang)
    e.style.display = 'block';
    e.autoplay = 'true';
    return false;
}
function changeLang() {
    var e = document.getElementById('audio')
    if (e.style.display == 'block') {
        playAudio();
    }
}
function reload() {
    setSrcQuery(document.getElementById('image'), "reload=" + (new Date()).getTime());
    setSrcQuery(document.getElementById('audio'), (new Date()).getTime());
    return false;
}
</script>
<form action="/process" method=post>
<span style="color: #595959; font-family: 楷体, 楷体_GB2312, SimKai; font-size: 24px;">用户名</span><br/>
<input name="user" type="text"/><br/>
<span style="color: #595959; font-family: 楷体, 楷体_GB2312, SimKai; font-size: 24px;">密码<br/></span>
<input name="usertext" type="password"/><br/>
&nbsp; &nbsp; &nbsp; &nbsp;
<p>
    &nbsp; &nbsp;
</p>
<p><img id=image src="/captcha/{{.CaptchaId}}.png" alt="Captcha image"></p>
<a href="#" onclick="reload()">Reload</a> | <a href="#" onclick="playAudio()">Play Audio</a>
<audio id=audio controls style="display:none" src="/captcha/{{.CaptchaId}}.wav" preload=none>
  You browser doesn't support audio.
  <a href="/captcha/download/{{.CaptchaId}}.wav">Download file</a> to play it in the external player.
</audio>
<input type=hidden name=captchaId value="{{.CaptchaId}}"><br>
<input name=captchaSolution>
<p>
    <input type="submit" value="登录||注册||游客" style="width:100px;background-color:grey"/>
</p>
</form>
`

const indexTr = `<html>
<head></head>
<body style="text-align:center;"> 
<form action="/loadgra" method=post>
<input type="submit" value="看图" name="dowhat" style="width:100px;"/>
<input type="hidden" value="游客" name="user";"/>
</form>
<form action="/loadtxt" method=post>
<input type="submit" value="看文" name="dowhat" style="width:100px;"/>
<input type="hidden" value="游客" name="user";"/>
</form>
<form action="/textdir" method=post>
<input type="submit" value="文录" name="dowhat" style="width:100px;"/>
<input type="hidden" value="游客" name="user";"/>
</form>
</body>
</html>
`
const indexAdmin = `<html>
<head></head>
<body style="text-align:center;"> 
<form action="/delgra" method=post>
<input type="submit" value="删图" name="dowhat" style="width:100px;"/>
<input type="hidden" value="admin" name="user";"/>
</form>
<form action="/deltext" method=post>
<input type="submit" value="删文" name="dowhat" style="width:100px;"/>
<input type="hidden" value="admin" name="user";"/>
</form>
<form action="/delcommend" method=post>
<input type="submit" value="删评" name="dowhat" style="width:100px;"/>
<input type="hidden" value="admin" name="user";"/>
</form>
<form action="/deluser" method=post>
<input type="submit" value="删人" name="dowhat" style="width:100px;"/>
<input type="hidden" value="admin" name="user";"/>
</form>
<form action="/addannouncement" method=post>
<input type="submit" value="公告" name="dowhat" style="width:100px;"/>
<input type="hidden" value="admin" name="user";"/>
</form>
</body>
</html>
`

// 用于将页面重定向到主页
const redirectHTML = `<html>
<head>
	<meta http-equiv="Content-type" content="text/html; charset=utf-8">
	<meta http-equiv="Refresh" content="0; url={{.}}">
</head>
<body></body>
</html>`

const uploadone = `<!doctype html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Document</title>
</head>
<body>
    <form action="/uploadtxt" method="post" enctype="multipart/form-data">
        文件：<input type="file" name="file" value="">
        <input type="submit" value="提交">
    </form>
</body>
</html>`

const uploadgraph = `<!doctype html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport"
          content="width=device-width, user-scalable=no, initial-scale=1.0, maximum-scale=1.0, minimum-scale=1.0">
    <meta http-equiv="X-UA-Compatible" content="ie=edge">
    <title>Document</title>
</head>
<body>
<form action="/uploadgra" method="post" enctype="multipart/form-data">
    图片：<br>
    <input type="file" name="img" value="请上传图片">
    <br><br>
    <input type="submit" value="submit">
</form>
</body>
</html>`

var indexHTML = `<html>
<head></head>
<body style="text-align:center;"> 
<form action="/uploadgra" method=post>
<input type="submit" value="交图" name="dowhat" style="width:100px;"/>
</form>
<form action="/loadgra" method=post>
<input type="submit" value="看图" name="dowhat" style="width:100px;"/>
</form>
<form action="/loadtxt" method=post>
<input type="submit" value="看文" name="dowhat" style="width:100px;"/>
</form>
<form action="/uploadtxt" method=post>
<input type="submit" value="交文" name="dowhat" style="width:100px;"/>
</form>
<form action="/myfriends" method=post>
<input type="submit" value="好友" name="dowhat" style="width:100px;"/>
</form>
<form action="/textdir" method=post>
<input type="submit" value="文录" name="dowhat" style="width:100px;"/>
</form>
<form action="/searchfiles" method=post>
<input type="submit" value="查找" name="dowhat" style="width:100px;"/>
</form>
</form>
<form action="/readannouncement" method=post>
<input type="submit" value="公告" name="dowhat" style="width:100px;"/>
</form>
</body>
</html>
`
var findTextDirHTML = `<html><form action="/searchfiles" method=post>
<input type="text" name="查找内容">
<input type="submit" value="查找" name="find" ;"/>
<select name="查找方式">
<option value="1">文章名</option>
<option value="2">文章内容</option>
<option value="3">发布日期</option>
<option value="4">文章作者</option>
</select></form><br><br>`
var addannouncement = `<html><body><form action="/addannouncement" method=post>
<p>标题：<input type="text" name="公告标题"></p>
<p>公告：<textarea cols="30" rows="5" name="公告内容"></textarea></p>
<input type="submit" name="提交">
</body></form><html>`
