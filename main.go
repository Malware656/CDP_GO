package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"time"
	"log"
	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/protocol/page"
	"github.com/mafredri/cdp/rpcc"
	"os/exec"
)

func main() {

	err := run(5 * time.Second)
	if err != nil {
		log.Fatal(err)
	}
	convToPdf(2);
}

func run(timeout time.Duration) error {
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Use the DevTools HTTP/JSON API to manage targets (e.g. pages, webworkers).
	devt := devtool.New("http://127.0.0.1:4000")
	pt, err := devt.Get(ctx, devtool.Page)
	if err != nil {
		pt, err = devt.Create(ctx)
		if err != nil {
			return err
		}
	}
	
	// Initiate a new RPC connection to the Chrome DevTools Protocol target.
	conn, err := rpcc.DialContext(ctx, pt.WebSocketDebuggerURL)
	if err != nil {
		return err
	}
	defer conn.Close() // Leaving connections open will leak memory.

	c := cdp.NewClient(conn)

	// Open a DOMContentEventFired client to buffer this event.
	domContent, err := c.Page.DOMContentEventFired(ctx)
	if err != nil {
		return err
	}
	defer domContent.Close()
	
	// Enable events on the Page domain, it's often preferrable to create
	// event clients before enabling events so that we don't miss any.
	if err = c.Page.Enable(ctx); err != nil {
		return err
	}

	// Create the Navigate arguments with the optional Referrer field set.
	navArgs := page.NewNavigateArgs("file:///home/calibraint/Desktop/template.html").
		SetReferrer("file:///home/calibraint/Desktop/template.html")
	_, err = c.Page.Navigate(ctx, navArgs)
	if err != nil {
		return err
	}

	// Wait until we have a DOMContentEventFired event.
	if _, err = domContent.Recv(); err != nil {
		return err
	}

	// Fetch the document root node. We can pass nil here
	// since this method only takes optional arguments.
	_, err = c.DOM.GetDocument(ctx, nil)
	if err != nil {
		return err
	}


	pdfName := "page.pdf"
	f, err := os.Create(pdfName)
	if err != nil {
		return err
	}

	pdfArgs := page.NewPrintToPDFArgs().
		SetTransferMode("ReturnAsStream") // Request stream.
	pdfData, err := c.Page.PrintToPDF(ctx, pdfArgs)
	if err != nil {
		return err
	}

	sr := c.NewIOStreamReader(ctx, *pdfData.Stream)
	r := bufio.NewReader(sr)

	// Write to file in ~r.Size() chunks.
	_, err = r.WriteTo(f)
	if err != nil {
		return err
	}

	err = f.Close()
	if err != nil {
		return err
	}
	
	end := time.Now()
	total := end.Sub(start).Milliseconds()
	fmt.Println(total, "ms")
	fmt.Printf("Saved PDF: %s\n", pdfName)

	return nil
}

func convToPdf(num int){
	start := time.Now()
	cmd := exec.Command("google-chrome", "--headless", "--print-to-pdf=report.pdf", "template.html")
	_, err := cmd.Output()
	if err != nil{
		fmt.Println(err.Error())
	}
	end := time.Now()
	total := end.Sub(start).Milliseconds()
	fmt.Println(total, "ms")
}