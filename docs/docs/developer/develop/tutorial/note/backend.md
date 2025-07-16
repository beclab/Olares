# Develop Backend Program

## Clone Code

  Open the IDE of backend Dev Container, open Terminal, clone your code to the `/Code` directory

  ```sh
  gh auth login

  cd /Code
  git clone https://github.com/beclab/terminus-app-demo.git
  ```
  
  You can then open the backend code in the IDE for development.
  
  ![server IDE](/images/developer/develop/tutorial/backend/dev.jpg)

## Connect Database

  In the Dev Container, you can access database details through environment variables. You can do this by adding the database parameters into the container using environment variables when you deploy it.

  Take `gorm` as an example:
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

## Debug

  After completing the development, you can run and debug your code in the IDE.

  ![run and debug](/images/developer/develop/tutorial/backend/debug.jpg)

  You can also run your code in the Terminal, for example:
  
  ```sh
  go run main.go
  ```
  
  Now, you can debug your interface with your front-end program.