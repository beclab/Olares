# 开始后端程序开发

## 克隆代码

  打开后端的开发容器 IDE, 打开 Terminal，可控你的代码到 `/Code` 目录。

  ```sh
  gh auth login

  cd /Code
  git clone https://github.com/beclab/terminus-app-demo.git
  ```

  之后便可以在 IDE 中打开后端代码进行开发。

  ![server IDE](/images/developer/develop/tutorial/backend/dev.jpg)

## 连接数据库

  在开发容器中，可以通过环境变量获取数据库信息（如果你在部署的时候以环境变量的方式将数据库参数注入容器）。

  以 gorm 为例：
  ```go
  import (
    "fmt"
    "os"
    "strconv"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
  )


  func init() {
    var err error

    db_host = os.Getenv("DB_HOST")
    db_port, err = strconv.Atoi(os.Getenv("DB_PORT"))
    if err != nil {
        panic(err)
    }
    db_username = os.Getenv("DB_USER")
    db_password = os.Getenv("DB_PWD")
    db_name = os.Getenv("DB_NAME")
  }


  func main(){
    dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
  	db_host, db_username, db_password, db_name, db_port)
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
  	  panic(err)
    }

  }
  ```

## 调试

  完成代码开发后，便可在 IDE 中运行调试你的代码。

  ![run and debug](/images/developer/develop/tutorial/backend/debug.jpg)

  也可以在 Terminal 中运行你的代码，例如：

  ```sh
  go run main.go
  ```

  这时，就可以配合前端完成接口联调。