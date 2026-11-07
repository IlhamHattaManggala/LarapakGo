package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/IlhamHattaManggala/larapak-go/pkg/ngapak"
)

const version = "1.0.0"

func main() {
	if len(os.Args) < 2 {
		startREPL()
		return
	}

	arg := os.Args[1]
	switch strings.ToLower(arg) {
	case "-v", "--version", "version", "versi":
		fmt.Printf("NgapakIn Language v%s (Go Edition)\n", version)

	case "-h", "--help", "help", "tulung":
		showHelp()

	case "repl", "interaktif":
		startREPL()

	case "run", "mlakukna", "eksekusi":
		if len(os.Args) < 3 {
			fmt.Println("Kliru: Kudune ana path berkas .ngpk. Contohe: ngapakin run skrip.ngpk")
			os.Exit(1)
		}
		runFile(os.Args[2])

	default:
		if strings.HasSuffix(arg, ".ngpk") || fileExists(arg) {
			runFile(arg)
		} else {
			fmt.Printf("Perintah ora dingerteni: %s\nKetik 'ngapakin tulung' nggo ndeleng bantuan.\n", arg)
			os.Exit(1)
		}
	}
}

func runFile(filePath string) {
	eng := ngapak.NewEngine(".", nil)
	_, err := eng.RunFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Kliru Eksekusi: %v\n", err)
		os.Exit(1)
	}
}

func startREPL() {
	fmt.Printf(`
   _  __                __   ____     
  / |/ /__ ____ ____  ___ _/ /__/  _/__ 
 /    / _ '/ _ '/ _ \/ _ '/ //_// / / _ \
/_/|_/\_, /\_,_/ .__/\_,_/\__,_//_/_/_//_/
     /___/    /_/                         
  NgapakIn Language REPL v%s (Go Edition)
  Ora Ngapak Ora Kepenak!
  Ketik 'metu' utawa pencet Ctrl+C nggo metu.

`, version)

	scanner := bufio.NewScanner(os.Stdin)
	eng := ngapak.NewEngine(".", nil)

	for {
		fmt.Print("ngapak>>> ")
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if line == "metu" || line == "exit" || line == "quit" {
			fmt.Println("Matur nuwun, rampung.")
			break
		}

		res, err := eng.RunSource(line, "repl")
		if err != nil {
			fmt.Printf("Kliru: %v\n", err)
		} else if res != nil {
			fmt.Println(res)
		}
	}
}

func showHelp() {
	fmt.Printf(`
NgapakIn Language CLI - Versi %s (Go Edition)
100%% Standard Library • Super Cepet & Hemat Memori

Cara Nganggo:
  ngapakin <berkas.ngpk>              Mlakukna skrip NgapakIn langsung
  ngapakin run <berkas.ngpk>          Mlakukna skrip NgapakIn
  ngapakin mlakukna <berkas.ngpk>     Mlakukna skrip (Dialek Ngapak)
  ngapakin repl                       Miwiti terminal interaktif REPL
  ngapakin -v / --version             Ndeleng versi NgapakIn
  ngapakin tulung / --help            Ndeleng panduan kiye

Contohe:
  ngapakin halo.ngpk
  ngapakin run dalan/web.ngpk
`, version)
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
