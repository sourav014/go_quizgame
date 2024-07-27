package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"
)

func parseFlags() (string, bool, int) {
	problemFileName := flag.String("problemFileName", "problems.csv", "Provide the problem file name.")
	shuffle := flag.Bool("shuffle", false, "Shuffle the problems.")
	timer := flag.Int("timer", 10, "Provider the timer for the Quiz.")
	flag.Parse()
	return *problemFileName, *shuffle, *timer
}

func getQuestionsAndAnswersFromCSVFIle(fileName string) ([]string, []string, error) {
	questions := []string{}
	answers := []string{}
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println("Error while opening the file.")
		return questions, answers, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error while reading the file.")
		return questions, answers, err
	}
	for _, record := range records {
		questions = append(questions, record[0])
		answers = append(answers, record[1])
	}
	return questions, answers, nil
}

func startQuiz(exitQuiz chan bool, questions []string, answers []string, quizResultStats chan int, wg *sync.WaitGroup) {
	defer wg.Done()

	reader := bufio.NewReader(os.Stdin)
	correctCount := 0
	wrongCount := 0
	totalCount := 0

	for i := 0; i < len(questions); i++ {
		select {
		case <-exitQuiz:
			fmt.Println("Times Up.")
			quizResultStats <- totalCount
			quizResultStats <- correctCount
			quizResultStats <- wrongCount
			return
		default:
			fmt.Printf("Problem %d\n", i+1)
			fmt.Printf("Question: %s\n", questions[i])
			fmt.Print("Please Enter the Answer: ")
			enteredAnswer, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error while reading the answer")
				quizResultStats <- totalCount
				quizResultStats <- correctCount
				quizResultStats <- wrongCount
				return
			}
			enteredAnswer = strings.TrimSpace(enteredAnswer)
			totalCount += 1
			if enteredAnswer == answers[i] {
				correctCount += 1
			} else {
				wrongCount += 1
			}
		}
	}
	quizResultStats <- totalCount
	quizResultStats <- correctCount
	quizResultStats <- wrongCount
}

func shuffleQuestions(questions []string) {
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	r.Shuffle(len(questions), func(i, j int) {
		questions[i], questions[j] = questions[j], questions[i]
	})
}

func startTimer(timeLimit int, exitQuiz chan bool, wg *sync.WaitGroup) {
	defer wg.Done()
	runDuration := timeLimit * int(time.Second)
	timer := time.NewTimer(time.Duration(runDuration))
	defer timer.Stop()
	<-timer.C
	exitQuiz <- true
}

func main() {
	problemFileName, shuffle, timer := parseFlags()
	questions, answers, err := getQuestionsAndAnswersFromCSVFIle(problemFileName)
	if err != nil {
		fmt.Println("Error while getQuestionsAndAnswersFromCSVFIle.")
		return
	}
	if shuffle {
		shuffleQuestions(questions)
	}
	reader := bufio.NewReader(os.Stdin)
	var wg sync.WaitGroup
	exitQuiz := make(chan bool)
	quizResultStats := make(chan int)

	fmt.Print("Press Enter to Start the Quiz! ")
	_, _ = reader.ReadString('\n')

	wg.Add(1)
	go startTimer(timer, exitQuiz, &wg)

	wg.Add(1)
	go startQuiz(exitQuiz, questions, answers, quizResultStats, &wg)

	go func() {
		wg.Wait()
		close(exitQuiz)
		close(quizResultStats)
	}()

	quizResult := []int{}
	for quizResultStat := range quizResultStats {
		quizResult = append(quizResult, quizResultStat)
	}

	fmt.Printf("Total Answers Count: %d\n", quizResult[0])
	fmt.Printf("Correct Answers Count: %d\n", quizResult[1])
	fmt.Printf("Wrong Answers Count: %d\n", quizResult[2])
}
