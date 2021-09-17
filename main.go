package main

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/jackc/pgx/v4"
)

/*
	Замечание 1. Заглавная буква в начале имени типа структуры (не критично, но можно начать со строчной)
	main пакеты не импортируются, поэтому экспорт идентификаторов из main пакетов не требуется.
	Не экспортируйте идентификаторы из main пакета, если вы собираете пакет в двоичный файл.
	https://golang-blog.blogspot.com/2020/06/package-style-in-golang.html
*/

type BookModel struct {
	Title  string
	Author string
	Cost   int
}

type Service struct {
	Pool   []*pgx.Conn
	IsInit bool
}

/*
	Issue 6. По поводу передачи username и password в функцию initService() - см. Issue 4
*/
func (s Service) initService(username, password string) {
	/*
		Issue 7. В коде ниже организован пул подключений к БД в горутине, но не организован процесс
		проверки, успеет ли горутина выполниться до использования пула подключений в функции getBooksByAuthor().
		Т.е. использовать асинхронную процедуру подключения к БД в синхронном выполнении кода не рекомендуется,
		если в синхронном коде нет уверенности в том, что асинхронный код успеет выполниться и вернуть результаты,
		используемые в синхронном коде.
		Рекомендация - убрать горутину, вытащив из нее код в текущий метод
	*/
	var backgroundTask = func() {
		/*
			Issue 8. Для удобства удобнее осуществлять конкатенацию строки через функцию fmt.Sprintf():
			var databaseUrl string = fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", user, password, host, port, dbname)
		*/
		var databaseUrl = "postgres://" + username + ":" + password + "@10.7.27.34:5432/books"

		/*
			Issue 9. Какую задачу выполняет цикл? Если речь про пул соединений, то см. реализацию https://eax.me/golang-pgx/
			Создавая подключения, необходимо помнить об отключении - defer conn.Close(context.Background())
		*/
		for i := 1; i <= 10; i++ {
			conn, err := pgx.Connect(context.Background(), databaseUrl)
			if err != nil {
				/*
					Issue 10. При обработке ошибок подключения к СУБД, желательно
					передавать err в качестве параметра, например:
					if err != nil {
						fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
						os.Exit(1)
					}
					Таким образом, можно узнать подробную информацию об ошибке. А в panic(nil) - не узнаем.
				*/
				println("Ошибка при подключении к базе по URL = " + databaseUrl)
				panic(nil)
			}
			s.Pool = append(s.Pool, conn)
		}
	}

	go backgroundTask()
}

/*
	Issue 3. Тип возвращаемого значения функции mux.Vars(r) есть map[string]string
	Для простоты использования и читабельности кода предлагается использовать переменную author
	в функции getBooksByAuthor() типа string, а не как указатель на *string. Ведь разыменование
	переменной - это тоже операция, на выполнение которой затрачивается квант процессорного времени
	https://github.com/gorilla/mux#examples

	Issue 4. При логическом разделении кода на блоки, можно выделить три части:
	-блок функций для работы с базой данных
	-блок обработчиков маршрутов (route handlers)
	-блок-точка входа в программу
	Следуя этой логике, необходимо относить переменные/константы/функции к соответствующим блокам.
	Параметры username, password в функции getBooksByAuthor() используются для подключения к БД, и
	не используются явно в самом методе. Кроме того, при добавлении новых маршрутов и их обработчиков,
	учитывая текущие сигнатуры функций, придется "тянуть" логин и пароль от БД в эти обработчики.
	Предлагаемое решение заключается в переносе информации о доступе к БД непосредственно в метод
	инициализации БД, либо вывести как константы в config файл. Таким образом, есть рекомендация
	убрать username, password из определения метода getBooksByAuthor(). Кроме того, username и password
	передаются в функцию как строки, а не хранятся в переменных окружения. Это может затруднить
	поиск и обновление кода при изменении username и password, что является менее универсальным способом
	Переменные окружения — лучший способ хранения конфигурации приложения, поскольку они могут быть заданы
	на системном уровне. Это один из принципов методологии Twelve-Factor App, он позволяет отделять приложения
	от системы, в которой они запущены (конфигурация может существенно различаться между деплоями, код не должен различаться).
	https://habr.com/ru/post/446468/
*/
func (s Service) getBooksByAuthor(username, password string, author *string, result []BookModel) {
	if !s.IsInit {
		s.initService(username, password)
		/*
			Issue 11. Чтобы изменить значение в "экземпляре" структуры s, необходимо передать указатель на структуру:
			func (s *Service) getBooksByAuthor()

			Параметры функции и метода передаются через значение. Это значит, что функции всегда оперируют копией
			переданных аргументов. Когда указатель передается функции, функция получает копию адреса памяти.
			Через разыменование адреса памяти функция может изменить значение, на которое указывает указатель.
			https://golangify.com/pointers#param
		*/
		s.IsInit = true
	}

	/*
		Issue 12. Выборка первого активного подключения к БД. Цикл не требуется, если создать одно подключение
		https://github.com/jackc/pgx
	*/
	var conn *pgx.Conn
	for _, x := range s.Pool {
		if !x.IsClosed() {
			conn = x
			break
		}
	}

	/*
		Issue 13. Форматирование строки можно реализовать с помощью параметров:
		conn.Query("SELECT * FROM album WHERE artist = ?", artist)
		https://golang.org/doc/database/querying#multiple_rows
	*/
	rows, err := conn.Query(context.Background(), "select title, cost from books where author="+*author)
	if err != nil {
		/*
			см. Issue 1
		*/
		println("Не удалось получить книги по автору")
		panic(nil)
	}

	for rows.Next() {
		var title string
		var cost int
		err = rows.Scan(&title, &cost)
		/*
			Issue 14. Необходимо добавить обработку ошибок, в случае, если err != nil
			https://golang.org/doc/database/querying#multiple_rows
		*/
		if err == nil {
			result = append(result, BookModel{title, *author, cost})
		}
	}

	/*
		см. Issue 1
	*/
	println("Успешно выполнен запрос, заполнено записей: " + strconv.Itoa(len(result)))
}

func main() {
	/*
		Issue 1. Мы можем использовать стандартный вывод fmt.Println() вместо println(), поскольку
		семейство, предоставляемое fmt, построено так, чтобы быть в production-коде. Они предсказуемо
		сообщают на стандартный вывод, если не указано иное. Они более универсальны (fmt.Fprint * может
		передавать отчеты любому io.Writer, например os.Stdout, os.Stderr или даже типу net.Conn.
		И не зависят от реализации.

		Большинство пакетов, отвечающих за вывод, имеют fmt в качестве зависимости, например log.
		Если ваша программа собирается выводить что-либо в производственной среде, fmt, скорее всего,
		будет тем пакетом, который вам нужен.
		https://stackoverflow.com/questions/14680255/difference-between-fmt-println-and-println-in-go
		https://golang.org/doc/effective_go#printing
	*/
	println("Запуск сервера...")
	var service = Service{}

	r := mux.NewRouter()
	/*
		Issue 2. Определение в качестве параметра анонимной функции и исполнения части кода в ней, с дальнейшим вызовом
		функции-обработчика getBooksByAuthor() может затруднять понимание кода.
		Мы можем изменить код таким образом, чтобы HandleFunc() содержал только URL путь и метод-обработчик этого пути, например:

		r.HandleFunc("/GetBookByAuthor/{author}", getBookByAuthor)

		Это позволит покрыть тестами код в теле анонимной функции (тестируя getBookByAuthor()) и работать с этим кодом в
		контексте функции getBookByAuthor()
		func getBookByAuthor(w http.ResponseWriter, r *http.Request) {
			vars := mux.Vars(r)
			author := vars["author"]
			...
		}
	*/
	r.HandleFunc("/GetBookByAuthor/{author}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		author := vars["author"]
		/*
			Issue 5. Переменная result содержит "обнуленный" массив длины 10 и возвращает срез, который ссылается на этот массив.
			"Обнуленный" означает, что все элементы среза (BookModel) содержат значения по умолчанию
			(BookModel{Title: "", Author: "", Cost: 0}).
			Поскольку далее по коду в функции getBooksByAuthor() идет заполнение с помощью result = append(result, ...),
			то первые 10 элементов среза будут "обнуленными" и не имеют смысла, поскольку append добавляет элемент к срезу (в конец).
			Рекомендуется убрать переменную result , определенную вне функции.
			Кроме того, здесь реализован подход, когда переменная для сбора результата функции напрямую передается в саму функцию.
			В данном примере имеет смысл определять result внутри функции, собирать в нее результат и возвращать (return) результат.
		*/
		var result = make([]BookModel, 10)
		service.getBooksByAuthor("boris", "qwerty", &author, result)
	})

	http.ListenAndServe(":8080", r)
}
