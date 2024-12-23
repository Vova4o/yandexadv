all: iter1 iter2 iter3 iter4 iter5 iter6 iter7 iter8 iter9

vet: 
	go vet -vettool=./.tools/statictest ./...

test:
	@echo "running tests"
	go test -coverprofile=coverage.out ./... && \
	go tool cover -html=coverage.out -o coverage.html && \
	open coverage.html

iter1:
	@echo "iter1 starting tests for first iteration"
	cd ./cmd/server/ && \
	go build -o server && \
	cd - && \
	chmod +x ./metricstest && \
	./metricstest -test.v -test.run=^TestIteration1$$ -binary-path=cmd/server/server

iter2:
	@echo "iter2 starting tests for second iteration"
	cd ./cmd/server/ && \
	go build -o server && \
	cd - && \
	cd ./cmd/agent/ && \
	go build -o agent && \
	cd - && \
	chmod +x ./metricstest && \
	./metricstest -test.v -test.run=^TestIteration2[AB]*$$ \
		-source-path=. \
		-agent-binary-path=cmd/agent/agent

iter3:
	@echo "iter3 starting tests for third iteration"
	cd ./cmd/server/ && \
	go build -o server && \
	cd - && \
	cd ./cmd/agent/ && \
	go build -o agent && \
	cd - && \
	chmod +x ./metricstest && \
	./metricstest -test.v -test.run=^TestIteration3[AB]*$$ \
		-source-path=. \
		-agent-binary-path=cmd/agent/agent \
		-binary-path=cmd/server/server

iter4:
	@echo "iter4 starting tests for fourth iteration"
	@SERVER_PORT="8080"; \
	ADDRESS="localhost:$${SERVER_PORT}"; \
	TEMP_FILE=$$(mktemp); \
	./metricstest -test.v -test.run=^TestIteration4$$ \
		-agent-binary-path=cmd/agent/agent \
		-binary-path=cmd/server/server \
		-server-port=$${SERVER_PORT} \
		-source-path=.

iter5:
	@echo "iter5 starting tests for fifth iteration"
	@SERVER_PORT=$$(python3 -c "import random; print(random.randint(8080, 8090))"); \
	ADDRESS="localhost:$${SERVER_PORT}"; \
	TEMP_FILE=$$(mktemp); \
	cd ./cmd/server/ && \
	go build -o server && \
	cd - && \
	cd ./cmd/agent/ && \
	go build -o agent && \
	cd - && \
	chmod +x ./metricstest && \
	./metricstest -test.v -test.run=^TestIteration5$$ \
		-agent-binary-path=cmd/agent/agent \
		-binary-path=cmd/server/server \
		-server-port=$${SERVER_PORT} \
		-source-path=.

iter6:
	@echo "iter6 starting tests for sixth iteration"
	@SERVER_PORT=8080 \
	ADDRESS="localhost:8080" \
	TEMP_FILE="/tmp/tempfile" \
	cd ./cmd/server/ && \
	go build -o server && \
	cd - && \
	cd ./cmd/agent/ && \
	go build -o agent && \
	cd - && \
	chmod +x ./metricstest && \
	./metricstest -test.v -test.run=^TestIteration6$$ \
		-agent-binary-path=cmd/agent/agent \
		-binary-path=cmd/server/server \
		-server-port=$${SERVER_PORT} \
		-source-path=.

iter7:
	@echo "iter7 starting tests for seventh iteration"
	@SERVER_PORT=8080 \
	ADDRESS="localhost:8080" \
	cd ./cmd/server/ && \
	go build -o server && \
	cd - && \
	cd ./cmd/agent/ && \
	go build -o agent && \
	cd - && \
	chmod +x ./metricstest && \
	./metricstest -test.v -test.run=^TestIteration7$$ \
		-binary-path=cmd/server/server \
		-agent-binary-path=cmd/agent/agent \
		-server-port=$${SERVER_PORT} \
		-source-path=.

iter8:
	@echo "iter8 starting tests for eighth iteration"
	@SERVER_PORT=8080 \
	ADDRESS="localhost:8080"; \
	cd ./cmd/server/ && \
	go build -o server && \
	cd - && \
	cd ./cmd/agent/ && \
	go build -o agent && \
	cd - && \
	chmod +x ./metricstest && \
	./metricstest -test.v -test.run=^TestIteration8$$ \
		-binary-path=cmd/server/server \
		-agent-binary-path=cmd/agent/agent \
		-server-port=$${SERVER_PORT} \
		-source-path=.

iter9:
	@echo "iter9 starting tests for nine iteration"
	@SERVER_PORT=8080 \
	ADDRESS="localhost:8080"; \
	TEMP_FILE="tempfile.json"; \
	cd ./cmd/server/ && \
	go build -o server && \
	cd - && \
	cd ./cmd/agent/ && \
	go build -o agent && \
	cd - && \
	chmod +x ./metricstest && \
	./metricstest -test.v -test.run=^TestIteration9$$ \
		-agent-binary-path=cmd/agent/agent \
		-binary-path=cmd/server/server \
		-file-storage-path=$$TEMP_FILE \
		-server-port=$$SERVER_PORT \
		-source-path=.

iter10:
	@echo "iter10 starting tests for ten iteration"
	@SERVER_PORT=8080 \
	ADDRESS="localhost:8080"; \
	TEMP_FILE="tempfile.json"; \
	cd ./cmd/server/ && \
	go build -o server && \
	cd - && \
	cd ./cmd/agent/ && \
	go build -o agent && \
	cd - && \
	chmod +x ./metricstest && \
    ./metricstest -test.v -test.run=^TestIteration10[AB]$$ \
		-agent-binary-path=cmd/agent/agent \
		-binary-path=cmd/server/server \
		-database-dsn='postgres://postgres:mypassword@localhost:5432/metrix?sslmode=disable' \
		-server-port=$$SERVER_PORT \
		-source-path=.

iter11:
	@echo "iter11 starting tests for eleven iteration"
	@SERVER_PORT=8080 \
	ADDRESS="localhost:8080"; \
	TEMP_FILE="tempfile.json"; \
	cd ./cmd/server/ && \
	go build -o server && \
	cd - && \
	cd ./cmd/agent/ && \
	go build -o agent && \
	cd - && \
	chmod +x ./metricstest && \
	./metricstest -test.v -test.run=^TestIteration11$$ \
		-agent-binary-path=cmd/agent/agent \
		-binary-path=cmd/server/server \
		-database-dsn='postgres://postgres:mypassword@localhost:5432/metrix?sslmode=disable' \
		-server-port=$$SERVER_PORT \
		-source-path=.

iter12:
	@echo "iter12 starting tests for twelve iteration"
	@SERVER_PORT=8080 \
	ADDRESS="localhost:8080"; \
	TEMP_FILE="tempfile.json"; \
	cd ./cmd/server/ && \
	go build -o server && \
	cd - && \
	cd ./cmd/agent/ && \
	go build -o agent && \
	cd - && \
	chmod +x ./metricstest && \
	./metricstest -test.v -test.run=^TestIteration12$$ \
		-agent-binary-path=cmd/agent/agent \
		-binary-path=cmd/server/server \
		-database-dsn='postgres://postgres:mypassword@localhost:5432/metrix?sslmode=disable' \
		-server-port=$$SERVER_PORT \
		-source-path=.

iter13:
	@echo "iter13 starting tests for twelve iteration"
	@SERVER_PORT=8080 \
	ADDRESS="localhost:8080"; \
	TEMP_FILE="tempfile.json"; \
	cd ./cmd/server/ && \
	go build -o server && \
	cd - && \
	cd ./cmd/agent/ && \
	go build -o agent && \
	cd - && \
	chmod +x ./metricstest && \
	./metricstest -test.v -test.run=^TestIteration13$$ \
		-agent-binary-path=cmd/agent/agent \
		-binary-path=cmd/server/server \
		-database-dsn='postgres://postgres:mypassword@localhost:5432/metrix?sslmode=disable' \
		-server-port=$$SERVER_PORT \
		-source-path=.

iter14:
	@echo "iter14 starting tests for forteen iteration"
	@SERVER_PORT=8080 \
	ADDRESS="localhost:8080"; \
	TEMP_FILE="tempfile.json"; \
	cd ./cmd/server/ && \
	go build -o server && \
	cd - && \
	cd ./cmd/agent/ && \
	go build -o agent && \
	cd - && \
	chmod +x ./metricstest && \
	./metricstest -test.v -test.run=^TestIteration14$$ \
            -agent-binary-path=cmd/agent/agent \
            -binary-path=cmd/server/server \
            -database-dsn='postgres://postgres:mypassword@localhost:5432/metrix?sslmode=disable' \
            -key="$${TEMP_FILE}" \
            -server-port=$$SERVER_PORT \
            -source-path=.

