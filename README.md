# code-review-task
## Junior golang dev task: make a code review
### https://goplay.space/#hMh2XyixXRC,23 (if not exists, see https://github.com/vitalg93/code-review-task/blob/main/main.go)

Carefully read the code and figure out how it works, understand what errors and flaws there are in it, think again and find comments that I did not find at the previous stage. Additionally, a corrected working code of the service has been added, which receives `JSON` containing books by the specified author at the URL `localhost:8080/books/{author}`. Books are loaded from the postgreSQL database

## Input parameters
`{author}` - string value of author's lastname

## Output
`JSON` - list of author's books in JSON (if exists, or `null`)

## Example
For example, if your PostgreSQL database contains this rows:      
![alt text](https://raw.githubusercontent.com/vitalg93/hello-world/main/db_books.jpg)    
then, the response to the request   
`localhost:8080/books/Толстой` will be    
```[{"id":1,"title":"Война и мир","author":"Толстой","cost":250,{"id":3,"title":"Юность","author":"Толстой","cost":150},{"id":4,"title":"Анна Каренина","author":"Толстой","cost":450}]```,   
and the request    
`localhost:8080/books/Бунин` will be `null`

## Quick start
+ Install Go (if you haven't already): https://golang.org/doc/tutorial/getting-started
+ Clone project from Github `git clone https://github.com/vitalg93/code-review-task.git`
+ Create PostgreSQL database `books` and create a table `books` in it (table structure see in Example section). 
+ Check database credentials in code (`DB_USERNAME`, etc.) to correct connection with PostgreSQL.
+ Fill some table rows
+ In project directory run in terminal: `go run main.go`
+ Ready! Try to send URL in your browser `localhost:8080/books/{author}`