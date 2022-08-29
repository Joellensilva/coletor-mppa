package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/cdproto/browser"
	"github.com/chromedp/chromedp"
)

type crawler struct {
	// Aqui temos os atributos e métodos necessários para realizar a coleta dos dados
	downloadTimeout  time.Duration
	generalTimeout   time.Duration
	timeBetweenSteps time.Duration
	year             string
	month            string
	output           string
}

var selectedMonth, selectedYear string

func (c crawler) crawl() ([]string, error) {
	// Chromedp setup.
	log.SetOutput(os.Stderr) // Enviando logs para o stderr para não afetar a execução do coletor.
	alloc, allocCancel := chromedp.NewExecAllocator(
		context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/77.0.3830.0 Safari/537.36"),
			chromedp.Flag("headless", true), // mude para false para executar com navegador visível.
			chromedp.NoSandbox,
			chromedp.DisableGPU,
		)...,
	)
	defer allocCancel()

	// Criando o contexto do Chromedp
	ctx, cancel := chromedp.NewContext(
		alloc,
		chromedp.WithLogf(log.Printf), // remover comentário para depurar
	)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, c.generalTimeout)
	defer cancel()

	// NOTA IMPORTANTE: os prefixos dos nomes dos arquivos tem que ser igual
	// ao esperado no parser MPPA.

	// Realiza o download
	// Contracheque
	log.Printf("Realizando seleção (%s/%s)...", c.month, c.year)
	if err := c.selecionaContracheque(ctx, c.year, c.month); err != nil {
		log.Fatalf("Erro no setup:%v", err)
	}
	log.Printf("Seleção realizada com sucesso!\n")
	// cqFname := c.downloadFilePath("contracheques")
	// log.Printf("Fazendo download do contracheque (%s)...", cqFname)
	// if err := c.exportaPlanilha(ctx, cqFname); err != nil {
	// 	log.Fatalf("Erro fazendo download do contracheque: %v", err)
	// }
	// log.Printf("Download realizado com sucesso!\n")

	// Indenizações
	log.Printf("Realizando seleção (%s/%s)...", c.month, c.year)
	if err := c.selecionaIndenizacoes(ctx, c.year, c.month); err != nil {
		log.Fatalf("Erro no setup:%v", err)
	}
	log.Printf("Seleção realizada com sucesso!\n")
	iFname := c.downloadFilePath("indenizacoes")
	log.Printf("Fazendo download das indenizações (%s)...", iFname)
	if err := c.exportaPlanilha(ctx, iFname); err != nil {
		log.Fatalf("Erro fazendo download dos indenizações: %v", err)
	}
	log.Printf("Download realizado com sucesso!\n")

	//return []string{cqFname, iFname}, nil
	return []string{"ttt"}, nil
}

// Retorna os caminhos completos dos arquivos baixados.
func (c crawler) downloadFilePath(prefix string) string {
	return filepath.Join(c.output, fmt.Sprintf("membros-ativos-%s-%s-%s.xls", prefix, c.month, c.year))
}

func (c crawler) selecionaContracheque(ctx context.Context, year string, month string) error {
	//var err error
	monthMap := map[string]string{
		"01": "Janeiro",
		"02": "Fevereiro",
		"03": "Março",
		"04": "Abril",
		"05": "Maio",
		"06": "Junho",
		"07": "Julho",
		"08": "Agosto",
		"09": "Setembro",
		"10": "Outubro",
		"11": "Novembro",
		"12": "Dezembro",
	}
	chromedp.Run(ctx,
		// Acessando o site
		chromedp.Navigate("http://transparencia.mppa.mp.br/index.htm"),
		chromedp.Sleep(c.timeBetweenSteps),
		// Selecionando a opção contracheque
		chromedp.Click(`//*[@id="16"]/div[2]/button`, chromedp.BySearch, chromedp.NodeVisible),
		chromedp.Sleep(c.timeBetweenSteps),
	)
	// Verifica qual ano e qual mês já está selecionado no site para não selecionar novamente.
	// selectedMonth, err = c.getSelectedMonth(ctx)
	// if err != nil {
	// 	log.Fatalf("erro ao obter mês selecionado no site: %v", err)
	// }
	// selectedYear, err = c.getSelectedYear(ctx)
	// if err != nil {
	// 	log.Fatalf("erro ao obter ano selecionado no site: %v", err)
	// }
	// // Seleciona o ano
	// if selectedYear != year {
	// 	log.Printf("Selecionando o ano...")
	// 	selectYear := fmt.Sprintf(`//*[@id="DS"]/div/div/div[1]/div[@title="%s"]/div[1]`, year)
	// 	if err := chromedp.Run(ctx,
	// 		chromedp.Click(`//*[@id="50"]/div[2]/button`, chromedp.BySearch, chromedp.NodeVisible),
	// 		chromedp.Sleep(c.timeBetweenSteps),
	// 		chromedp.DoubleClick(selectYear, chromedp.BySearch, chromedp.NodeVisible),
	// 		chromedp.Sleep(c.timeBetweenSteps),
	// 	); err != nil {
	// 		return fmt.Errorf("Erro: %w", err)
	// 	}
	// }
	// Seleciona o mês
	//if selectedMonth != monthMap[month] {
	log.Printf("Selecionando o mês...")
	selectMonth := fmt.Sprintf(`//*[@id="DS"]/div/div/div[1]/div[@title="%s"]/div[1]`, monthMap[month])
	if err := chromedp.Run(ctx,
		chromedp.Click(`//*[@id="49"]/div[2]/button`, chromedp.BySearch, chromedp.NodeVisible),
		chromedp.Sleep(c.timeBetweenSteps),
		chromedp.DoubleClick(selectMonth, chromedp.BySearch, chromedp.NodeVisible),
		chromedp.Sleep(c.timeBetweenSteps),
	); err != nil {
		return fmt.Errorf("Erro: %w", err)
	}
	//}
	// Altera o diretório de download
	if err := chromedp.Run(ctx,
		browser.SetDownloadBehavior(browser.SetDownloadBehaviorBehaviorAllowAndName).
			WithDownloadPath(c.output).
			WithEventsEnabled(true),
	); err != nil {
		return fmt.Errorf("Erro: %w", err)
	}
	return nil
}

func (c crawler) selecionaIndenizacoes(ctx context.Context, year string, month string) error {
	monthMap := map[string]string{
		"01": "Janeiro",
		"02": "Fevereiro",
		"03": "Março",
		"04": "Abril",
		"05": "Maio",
		"06": "Junho",
		"07": "Julho",
		"08": "Agosto",
		"09": "Setembro",
		"10": "Outubro",
		"11": "Novembro",
		"12": "Dezembro",
	}
	var buf []byte

	chromedp.Run(ctx,
		// Seleciona a opção Verbas Indenizatórias e Outras Remunerações Temporárias
		chromedp.Click(`//*[@id="38"]/div[2]/button`, chromedp.BySearch, chromedp.NodeVisible),
		chromedp.Sleep(c.timeBetweenSteps),
		chromedp.WaitVisible(`//*[@id="111"]`, chromedp.BySearch),
		chromedp.FullScreenshot(&buf, 90),
	)
	if err := ioutil.WriteFile("/output/fullScreenshot.jpeg", buf, 0644); err != nil {
		log.Fatal(err)
	}
	print("eu to aqui")
	var err error
	selectedMonth, err = c.getSelectedMonth(ctx)
	if err != nil {
		log.Fatalf("erro ao obter mês selecionado no site: %v", err)
	}
	print("agora aqui")
	// selectedYear, err = c.getSelectedYear(ctx)
	// if err != nil {
	// 	log.Fatalf("erro ao obter ano selecionado no site: %v", err)
	// }
	// Seleciona o ano
	// if selectedYear != year {
	// 	log.Printf("Selecionando o ano...")
	// 	selectYear := fmt.Sprintf(`//*[@id="DS"]/div/div/div[1]/div[@title="%s"]/div[1]`, year)
	// 	if err := chromedp.Run(ctx,
	// 		chromedp.Click(`//*[@id="105"]/div[2]/button`, chromedp.BySearch, chromedp.NodeVisible),
	// 		chromedp.Sleep(c.timeBetweenSteps),
	// 		chromedp.DoubleClick(selectYear, chromedp.BySearch, chromedp.NodeVisible),
	// 		chromedp.Sleep(c.timeBetweenSteps),
	// 	); err != nil {
	// 		return fmt.Errorf("Erro: %w", err)
	// 	}
	// }
	// Seleciona o mês
	if selectedMonth != monthMap[month] {
		log.Printf("Selecionando o mês...")
		selectMonth := fmt.Sprintf(`//*[@id="DS"]/div/div/div[1]/div[@title="%s"]/div[1]`, monthMap[month])
		if err := chromedp.Run(ctx,
			chromedp.Click(`//*[@id="106"]/div[2]/button`, chromedp.BySearch, chromedp.NodeVisible),
			chromedp.Sleep(c.timeBetweenSteps),
			chromedp.DoubleClick(selectMonth, chromedp.BySearch, chromedp.NodeVisible),
			chromedp.FullScreenshot(&buf, 90),

			chromedp.Sleep(c.timeBetweenSteps),
		); err != nil {
			return fmt.Errorf("Erro: %w", err)
		}
		if err := ioutil.WriteFile("/output/fullScreenshot.jpeg", buf, 0644); err != nil {
			log.Fatal(err)
		}
	}
	// Altera o diretório de download
	if err := chromedp.Run(ctx,
		browser.SetDownloadBehavior(browser.SetDownloadBehaviorBehaviorAllowAndName).
			WithDownloadPath(c.output).
			WithEventsEnabled(true),
	); err != nil {
		return fmt.Errorf("Erro: %w", err)
	}
	return nil
}

// A função exportaPlanilha clica no botão correto para exportar para excel, espera um tempo para o download e renomeia o arquivo.
func (c crawler) exportaPlanilha(ctx context.Context, fName string) error {
	// Clica no botão de download
	if strings.Contains(fName, "contracheques") {
		// Contracheque
		chromedp.Run(ctx,
			chromedp.Click(`//*[@id="34"]/div[1]/div[1]/div`, chromedp.BySearch, chromedp.NodeVisible),
			chromedp.Sleep(c.downloadTimeout),
		)
	} else {
		// Indenizações
		chromedp.Run(ctx,
			chromedp.Click(`//*[@id="111"]/div[1]/div[1]`, chromedp.BySearch, chromedp.NodeVisible),
			chromedp.Sleep(c.downloadTimeout),
		)
	}

	if err := nomeiaDownload(c.output, fName); err != nil {
		return fmt.Errorf("erro renomeando arquivo (%s): %v", fName, err)
	}
	if _, err := os.Stat(fName); os.IsNotExist(err) {
		return fmt.Errorf("download do arquivo de %s não realizado", fName)
	}
	return nil
}

// A função nomeiaDownload dá um nome ao último arquivo modificado dentro do diretório
// passado como parâmetro
func nomeiaDownload(output, fName string) error {
	// Identifica qual foi o último arquivo
	files, err := os.ReadDir(output)
	if err != nil {
		return fmt.Errorf("erro lendo diretório %s: %v", output, err)
	}
	var newestFPath string
	var newestTime int64 = 0
	for _, f := range files {
		fPath := filepath.Join(output, f.Name())
		fi, err := os.Stat(fPath)
		if err != nil {
			return fmt.Errorf("erro obtendo informações sobre arquivo %s: %v", fPath, err)
		}
		currTime := fi.ModTime().Unix()
		if currTime > newestTime {
			newestTime = currTime
			newestFPath = fPath
		}
	}
	// Renomeia o último arquivo modificado.
	if err := os.Rename(newestFPath, fName); err != nil {
		return fmt.Errorf("erro renomeando último arquivo modificado (%s)->(%s): %v", newestFPath, fName, err)
	}
	return nil
}

/* funções getSelectedMonth e getSelectedYear

Procuram o valor que está no atributo "placeholder" do input de selecionar mês e ano.
Caso não encontrem o input, um erro será lançado.
Caso encontrem o input, mas não encontrem o atributo "placeholder" nele, a variável "ok" será igual a false.
Por fim, se encontrarem o input e ele possuir o atributo "placeholder", as variáveis
"selectedMonth/selectedYear" receberão os valores desses atributos.
*/
func (c crawler) getSelectedMonth(ctx context.Context) (string, error) {
	var ok bool
	var selectedMonth string

	if err := chromedp.Run(ctx,
		chromedp.AttributeValue(`//*[@id="106"]/div[2]/input`, "placeholder", &selectedMonth, &ok, chromedp.BySearch),
		chromedp.Sleep(c.timeBetweenSteps),
	); err != nil {
		return "", fmt.Errorf(`Erro recuperando valor: %w`, err)
	}

	//Verifica se o atributo "placeholder" foi encontrado dentro do input selecionado.
	if !ok {
		return "", fmt.Errorf(`O input selecionado não possui o atributo "placeholder"`)
	}

	return selectedMonth, nil
}

// func (c crawler) getSelectedYear(ctx context.Context) (string, error) {
// 	var ok bool
// 	var selectedYear string

// 	if err := chromedp.Run(ctx,
// 		chromedp.AttributeValue(`//*[@id="105"]/div[2]/input`, "placeholder", &selectedYear, &ok, chromedp.BySearch),
// 		chromedp.Sleep(c.timeBetweenSteps),
// 	); err != nil {
// 		return "", fmt.Errorf(`Erro recuperando valor: %w`, err)
// 	}

// 	//Verifica se o atributo "placeholder" foi encontrado dentro do input selecionado.
// 	if !ok {
// 		return "", fmt.Errorf(`O input selecionado não possui o atributo "placeholder"`)
// 	}

// 	return selectedYear, nil
// }
