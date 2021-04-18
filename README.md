<div>
    <h1>
		Kraicklist
    </h1>
    <p>a Search Ads Application</p>
    <p>
      <a href="https://github.com/isdzulqor/kraicklist/actions?query=workflow%3A%22Build%22">
          <img src="https://github.com/isdzulqor/kraicklist/workflows/Build/badge.svg?branch=master"/>
      </a>
      <a href="https://github.com/isdzulqor/kraicklist/actions?query=workflow%3A%22Test+Integration%22">
        <img src="https://github.com/isdzulqor/kraicklist/workflows/Test Integration/badge.svg?branch=master"/>
		  </a>
    </p>
</div>


- [Live Version](#live-version)
- [Features](#features)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Usage](#usage)
- [[DRAFT] Future Enhancements](#draft-future-enhancements)

## Live Version
- http://still-in-todo.com - with Elastic Search
- http://still-in-todo.com - with Bleve Search

## Features
- Full-text search with tf-idf from Bleve Search
- Fuzziness search with Elastic Search 
- Horizontal scalable supported
  - Graceful shutdown & healtcheck set up
- A correlation ID support in mind
- CI/CD: build, integration test, release with Github Action
  - release: Auto create changelog
  - release: Auto deploy to heroku & release to docker hub

## Prerequisites
- Golang 1.15^ - https://golang.org/dl/
- Run with Docker 
  - Docker - https://docs.docker.com/engine/install/
  - Docker Compose - https://docs.docker.com/compose/install/
- Run with Elastic Search
  - https://www.elastic.co/guide/en/elasticsearch/reference/current/install-elasticsearch.html

## Installation
- Install [the prerequisites](#prerequisites)
- Use Indexer with [Bleve Search](http://blevesearch.com/)
  - Environment variables need to set up and/or overwrite
    ```
    INDEXER_ACTIVATED=bleve
    ADVERTISEMENT_BLEVE_INDEX_NAME=kraicklist.bleve
    ```
- Use Indexer with [Elastic Search](https://www.elastic.co//)
  - Environment variables need to set up and/or overwrite
    ```
    INDEXER_ACTIVATED=elastic
    ELASTIC_HOST=http://localhost:9200
    ELASTIC_USERNAME=elastic
    ELASTIC_PASSWORD=elastic-password
    ```
- Visit http://localhost:7000 for the UI

## Quick Start
- Use `go run` command
  ```
  # run data seeding for first initiation
  $ go run main.go seed

  # start http server
  $ go run main.go api
  ```
- Use Makefile
  - Modify configs on [.env](.env) file
  - Makefile commands
    ```
    # will cleanup existing index, seed data for first initiation, and start http server
    $ make all

    # drop the created index
    $ make clean-index

    # start http server
    $ make dev
    
    # seed data with data from data.gz
    $ make seed
    
    # running integration test with some scenarios
    $ make integration-test
    ```
- Use Docker
  ```
  # replace the YOUR_PATH_DATA with PATH on your host machine
  $ docker run --env-file .env -p 7000:7000 -v YOUR_PATH_DATA:/data isdzulqor/kraicklist:latest sh -c "./app seed && ./app api"
  ```
- Use Docker Compose
  ```
  $ docker-compose up -d
  ```
## Usage
- Search ads
  ```
  $ curl --location --request GET 'http://localhost:7000/api/advertisement/search?q=iphone%20cheap'
  ```
- Index new ads 
  ```
  $ curl --location --request POST 'http://localhost:7000/api/advertisement/index' \
  --header 'Content-Type: application/json' \
  --data-raw '[
      {
          "id": 63983811,
          "title": "M attorney and legal advisor for drafting objection regulations",
          "content": "We do our best to achieve justice and strive to achieve the goal. To communicate and consider any case or advice\n##### We serve all regions of the Kingdom #####\nA group of lawyers and legal consultants with sufficient experience and expertise in all laws and regulations\nWe study all cases and provide legal advice that falls within the jurisdiction of our offices.\nFiling all types of lawsuits electronically (legal, commercial, administrative, personal status, labor)\n** Drafting regulations, answer notes, objection regulations, and a request for reconsideration.\nWriting contracts of all kinds between individuals",
          "thumb_url": "https://imgcdn.haraj.com.sa/cache2/600x338-1_-5f789b1619c4f.jpeg",
          "tags": [
              "قسم غير مصنف",
              "كل الحراج"
          ],
          "updated_at": 1616050291,
          "image_urls": [
              "https://mimg1cdn.haraj.com.sa/userfiles30/2020-10-03/600x338-1_-5f789b1619c4f.jpeg",
              "https://mimg1cdn.haraj.com.sa/userfiles30/2020-10-03/650x877-1_-5f789b1713147.jpeg",
              "https://mimg1cdn.haraj.com.sa/userfiles30/2020-10-03/300x400-1_-5f789b17aa4fe.jpeg"
          ]
      }
  ]'
  ```
- Health check
  ```
  $ curl --location --request GET 'http://localhost:7777/health' --header 'x-health-token: health-token'
  ```

## [DRAFT] Future Enhancements
- Product Side
  - Conduct user behavior analysis
    - Add event tracking on Front End
  - Infinite scolling/pagination
  - Most searched placeholders
  - Auto suggestions
  - UI/UX & Information completeness
- System Side
  - Revisit scoring system based on user behaviour analytics result
  - In-Mem cache for improving performance
  - Continuous profiling