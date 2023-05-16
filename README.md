# Simple-linebot

## Getting started

1. Run `go mod tidy`.

2. Create `.env` file from `example.env` and set values and secrets to `.env`.

3. Source the `.env` file.

    ```bash
    set -a && . .env && set +a
    ```

4. Start MongoDB.

    ```bash
    sudo -E ./run-mongo.sh
    ```

5. Start `ngrok` (Need to install before running)

    ```bash
    ngrok http 8080
    ```

6. Run the program.

    ```bash
    go run samplelinebot
    ```


---

* Stop MongoDB

    ```bash
    sudo -E ./run-mongo.sh stop
    ```

* Remove MongoDB

    ```bash
    sudo -E ./run-mongo.sh remove
    ```
