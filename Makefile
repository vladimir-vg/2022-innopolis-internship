node_deps:
	cd webui && npm install .

webui/public/sql-wasm.wasm: node_deps
	cp webui/node_modules/sql.js/dist/sql-wasm.wasm webui/public/

run_server: webui/public/sql-wasm.wasm
	cd webui && npm run start

.PHONY: node_deps run_server