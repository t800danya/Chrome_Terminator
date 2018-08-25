/*
 * (c) 2018 Козырев Богдан <t800@kvkozyrev.org>
 *
 */

package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"log"
	"os/exec"
        "path/filepath"
        "time"
        "syscall"
)

        //  Здесь задаем сколько свобдной памяти должно быть в сисеме,

       const mem int64 = 100000000  // Значение mem >=  0 в Килобайтах

       //  Здесь указваем команду которая запускает Chrome

       const runchrome string = "google-chrome"

      //  Здесь указваем имя процеса Chrome

       const appchrome string = "chrome"

     // Задержка между проверками памяти в миллисекундах

       const milisec time.Duration = 10000




// Функция проверка сколько свободной памяти в Системе парсингом meminfo

func FreeMem() int64 {
	filename := "/proc/meminfo"
	FileBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("Ошибка! -  не могу прочитать %s\n", filename)
		return -6
	}
	bufr := bytes.NewBuffer(FileBytes)
	for {
		line, err := bufr.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Sprintf("Ошибка! -  не могу прочитать %s\n", filename)
		}
		ndx := strings.Index(line, "MemFree:")
		if ndx >= 0 {
			line = strings.TrimSpace(line[9:])
			line = line[:len(line)-3]
			mem, err := strconv.ParseInt(line, 10, 64)
			if err == nil {
				return mem
			}
			// Обработка ошибок
			fmt.Printf("Строка: %s\n", line)
			n, err := fmt.Sscan(line, "%d", &mem)
			if err != nil {
				fmt.Printf("Ошибка! -  не могу просканировать %s\n", line)
				return -2
			}
			if n != 1 {
				fmt.Printf("Ошибка! -  не могу все просканировать на  %s\n", line)
				return -3
			}
			return -4
		}
	}
	fmt.Printf("Не могу найт FreeMem в /proc/meminfo\n")
	return -5
}


// Функция поиска процесса и его Убийства! 


func findAndKillProcess(path string, info os.FileInfo, err error) error {
    // 
    // Тут типа на случай ошибки с привелегиями. 
    //
    if err != nil {
        // log.Println(err)
        return nil
    }

    // Ищем файлы вида /proc/<pid>/status.
    if strings.Count(path, "/") == 3 {
        if strings.Contains(path, "/status") {

            // Извелкаем центральную часть и конвертим в pid
            // 
            pid, err := strconv.Atoi(path[6:strings.LastIndex(path, "/")])
            if err != nil {
                log.Println(err)
                return nil
            }


            //
            // Ошибка если не можем прочитать файл с именем процесса.
            //
            f, err := ioutil.ReadFile(path)
            if err != nil {
                log.Println(err)
                return nil
            }

            // Извлекаем имя процесса и буфера 

            name := string(f[6:bytes.IndexByte(f, '\n')])

            if name == appchrome {
                fmt.Printf("PID: %d, %s сейчас будет Убит!\n", pid, name)
                proc, err := os.FindProcess(pid)
                if err != nil {
                    log.Println(err)
                }
                // Убиваев Процесс! :-/
                proc.Kill()

                // Тут кроче завершаем поиск
                //
                return io.EOF
            }

        }
    }

    return nil
}



//  Функция запуска Chrome с Костылями для Убийства


func RunChrome() {
        for {
run:
      fmt.Printf("Запускаем Chrome \n")
      cmd := exec.Command(runchrome)
      if err := cmd.Start(); err != nil {
		print(err.Error())
		os.Exit(1)
	}
kill:
        if mem > FreeMem() {
        fmt.Println("Памяти не хватает")
	fmt.Println("Убиваем Chrome ",syscall.Kill(cmd.Process.Pid, syscall.SIGKILL))
        time.Sleep(time.Millisecond * milisec)
        goto run
	}
        if mem <= FreeMem() {
        time.Sleep(time.Millisecond * milisec)
        fmt.Println("Памяти хватает")
        fmt.Println("Не Убиваем Chrome ")
        goto kill
        }

   }
}



func main() {
              for {
                   fmt.Println("Проверка в ", time.Now())
                   time.Sleep(time.Millisecond * milisec)
                   fmt.Printf("Cвободная память %d\n",FreeMem())
                   if mem <= FreeMem() {
                   fmt.Println("Памяти хватает")
                     }  else {

                     fmt.Println("Памяти не хватает\n")
                     fmt.Printf("Пытаемся убить процесс \"%s\"\n", appchrome)
                     err := filepath.Walk("/proc", findAndKillProcess)
                     if err != nil {
                     if err == io.EOF {
                    // На ошибка а просто сигнал что дела сделано :)
                       err = nil
                      } else {
                      log.Fatal(err)
                       }
                     }
                     time.Sleep(time.Millisecond * milisec)
                     RunChrome()
                    }
                }

}
