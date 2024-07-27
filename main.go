package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	flagFilePath string
	flagRandom   bool
	flagTimer    int
	wg           sync.WaitGroup
)

func init() {
	flag.StringVar(&flagFilePath, "file", "problems.csv", "Provide the problem file name.")
	flag.BoolVar(&flagRandom, "flagRandom", true, "flagRandom the problems.")
	flag.IntVar(&flagTimer, "flagTimer", 10, "Provider the flagTimer for the Quiz.")
	flag.Parse()
}

func main() {
	file, err := os.Open(flagFilePath)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalln(err)
	}

	totalQuestions := len(records)
	questions := make(map[int]string, totalQuestions)
	answers := make(map[int]string, totalQuestions)
	responses := make(map[int]string, totalQuestions)

	for i, record := range records {
		questions[i] = record[0]
		answers[i] = record[1]
	}
	answer := make(chan string)

	fmt.Print("Press [Enter] to Start the Quiz!\n")
	bufio.NewScanner(os.Stdout).Scan()
	if flagRandom {
		rand.Seed(time.Now().UTC().UTC().UnixNano())
	}
	randPool := rand.Perm(totalQuestions)

	timeUp := time.After(time.Second * time.Duration(flagTimer))
	wg.Add(1)

	go func() {
	label:
		for i := 0; i < len(questions); i++ {
			index := randPool[i]
			go askQuesion(os.Stdout, os.Stdin, questions[index], answer)
			select {
			case <-timeUp:
				fmt.Fprint(os.Stdout, "\nTimes Up.")
				break label
			case ans, ok := <-answer:
				if ok {
					responses[index] = ans
				} else {
					break label
				}
			}
		}
		wg.Done()
	}()
	wg.Wait()

	correct := 0
	for i := 0; i < len(questions); i++ {
		if checkAnswer(answers[i], responses[i]) {
			correct++
		}
	}
	fmt.Fprintf(os.Stdout, "\nYou have answered %d questions correctly (%d / %d)\n", correct, correct, totalQuestions)
}

func askQuesion(w io.Writer, r io.Reader, quesion string, answer chan string) {
	reader := bufio.NewReader(r)
	fmt.Fprintf(w, "Question: %s\n", quesion)
	fmt.Fprintf(w, "Please Enter the Answer: ")
	enteredAnswer, err := reader.ReadString('\n')
	if err != nil {
		close(answer)
		if err == io.EOF {
			return
		}
		log.Fatalln(err)
	}
	enteredAnswer = strings.TrimSpace(enteredAnswer)
	answer <- enteredAnswer
}

func checkAnswer(ans string, expected string) bool {
	return strings.EqualFold(strings.TrimSpace(ans), strings.TrimSpace(expected))
}
