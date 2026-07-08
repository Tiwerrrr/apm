package main

import (
	"log"
	"net/http"
)

func main() {
	port := "8080"
	log.Printf("🚀 Сайт запущен локально на http://localhost:%s", port)
	
	// Раздаем статические файлы из текущей директории
	err := http.ListenAndServe(":"+port, http.FileServer(http.Dir(".")))
	if err != nil {
		log.Fatal("Ошибка запуска сервера: ", err)
	}
}
