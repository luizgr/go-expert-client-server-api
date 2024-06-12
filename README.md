# Client-Server API 

Este é um desafio proposto no curso de **GO Expert** onde temos duas aplicações **server** e **client** desenvolvidas na linguagem **GO**.

## Server

Esta aplicação deverá ser um servidor **HTTP** disponível na porta 8080, onde ao acessar o endpoint _/cotacao_, ela deverá fazer uma requisição no endereço [https://economia.awesomeapi.com.br/json/last/USD-BRL](https://economia.awesomeapi.com.br/json/last/USD-BRL) (para obter as informações de câmbio USD/BRL), gravar o **JSON** recebido em um banco de dados **SQLite** e retorná-lo através do write do **HTTP**. Ela também deverá utilizar contextos com tempo de execução limitados em **200ms** para requisição e **10ms** para persistência dos dados.

## Client

Esta aplicação deverá consumir o endpoint _/cotacao_ do **server** utilizando apenas o campo "bid" do **JSON** recebido e armazenar a informação em um arquivo de texto **cotacao.txt** no formato: _Dólar: {bid}_. O client também deverá utilizar contexto e limitar o tempo de execução em **300ms** na requisição.