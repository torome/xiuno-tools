package xn3ToXn4

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/skiy/golib"
	"log"
)

type user_open_plat struct {
	db3str,
	db4str dbstr
	fields user_open_platFields
}

type user_open_platFields struct {
	uid, platid, openid string
}

func (this *user_open_plat) update() {
	if !lib.AutoUpdate(this.db4str.Auto, this.db4str.DBPre+"user_open_plat") {
		return
	}

	count, err := this.toUpdate()
	if err != nil {
		log.Fatalln("转换 " + this.db3str.DBPre + "user_open_plat 失败: " + err.Error())
	}

	fmt.Printf("转换 %suser_open_plat 表成功，共(%d)条数据\r\n", this.db3str.DBPre, count)
}

func (this *user_open_plat) toUpdate() (count int, err error) {
	xn3pre := this.db3str.DBPre
	xn4pre := this.db4str.DBPre

	fields := "uid,platid,openid"
	qmark := this.db3str.FieldMakeQmark(fields, "?")
	xn3 := fmt.Sprintf("SELECT %s FROM %suser_open_plat", fields, xn3pre)
	xn4 := fmt.Sprintf("INSERT INTO %suser_open_plat (%s) VALUES (%s)", xn4pre, fields, qmark)

	createTable := `
CREATE TABLE IF NOT EXISTS %suser_open_plat (
	uid int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '用户编号',
	platid tinyint(1) NOT NULL DEFAULT '0' COMMENT '平台编号 0:本站 1:QQ 登录 2:微信登陆 3:支付宝登录 ',
	openid char(40) NOT NULL DEFAULT '' COMMENT '第三方唯一标识',
	PRIMARY KEY (uid),
	KEY openid_platid (platid,openid)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8
`

	data, err := xiuno3db.Query(xn3)
	if err != nil {
		log.Fatalln(xn3, err.Error())
	}
	defer data.Close()

	xn4CreateTable := fmt.Sprintf(createTable, xn4pre)
	_, err = xiuno4db.Exec(xn4CreateTable)
	if err != nil {
		log.Fatalln("Xiuno4: ", xn4CreateTable, err.Error())
	}

	xn4Clear := "TRUNCATE `" + xn4pre + "user_open_plat`"
	_, err = xiuno4db.Exec(xn4Clear)
	if err != nil {
		log.Fatalf(":::清空 %suser_open_plat 表失败: "+err.Error(), xn4pre)
	}
	fmt.Printf("清空 %suser_open_plat 表成功\r\n", xn4pre)

	tx, err := xiuno4db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare(xn4)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	fmt.Printf("正在升级 %suser_open_plat 表\r\n", xn4pre)

	var field user_open_platFields
	for data.Next() {
		err = data.Scan(
			&field.uid,
			&field.platid,
			&field.openid)

		_, err = stmt.Exec(
			&field.uid,
			&field.platid,
			&field.openid)

		if err != nil {
			fmt.Printf("导入数据失败(%s) \r\n", err.Error())
		} else {
			count++
		}
	}

	if err = data.Err(); err != nil {
		log.Fatalln(err.Error())
	}

	err = tx.Commit()
	if err != nil {
		log.Fatalln(err.Error())
	}

	return count, err
}
