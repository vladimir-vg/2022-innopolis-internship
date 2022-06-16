import { useEffect, useState } from 'react';

// import logo from './logo.svg';
import './App.css';
import initSqlJs from 'sql.js';

// Required to let webpack 4 know it needs to copy the wasm file to our assets
// import sqlWasm from "!!file-loader?name=sql-wasm-[contenthash].wasm!sql.js/dist/sql-wasm.wasm";

function SvgArea({ db }) {
  const [goroutines, setGoroutines] = useState([]);
  useEffect(() => {
    const rows = db.exec("SELECT * FROM goroutines");
    console.log(rows);
  }, [db]);

  return (
    <svg width="500" height="500">
      <rect x="10" y="10" width="100" height="100" />
    </svg>
  );
}


function App() {
  const [sql, setSql] = useState(null);
  const [db, setDb] = useState(null);

  useEffect(() => {
    const helper = async () => {
      const sql0 = await initSqlJs({
        locateFile: filename => `/${filename}`
      });
      setSql(sql0);
    }
    helper();
  }, []);

  const onFileSelect = (e) => {
    const f = e.target.files[0];
    const r = new FileReader();
    r.onload = function () {
      const bytes = new Uint8Array(r.result);
      setDb(new sql.Database(bytes));
    }
    r.readAsArrayBuffer(f);
  }

  const dbIsLoaded = !!db;
  // hide input, if db is created
  const fileInputDisplay = dbIsLoaded ? 'none' : 'block';

  return (
    <div className="App">
      <header style={{display: fileInputDisplay}} className="App-header">
        <input type="file" onChange={onFileSelect} />
      </header>
      {dbIsLoaded && <SvgArea db={db} />}
      {/* <header className="App-header">
        <img src={logo} className="App-logo" alt="logo" />
        <p>
          Edit <code>src/App.js</code> and save to reload.
        </p>
        <a
          className="App-link"
          href="https://reactjs.org"
          target="_blank"
          rel="noopener noreferrer"
        >
          Learn React
        </a>
      </header> */}
    </div>
  );
}

export default App;
