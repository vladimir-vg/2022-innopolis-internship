import { useEffect, useState } from 'react';

// import logo from './logo.svg';
import './App.css';
import initSqlJs from 'sql.js';

const CELL_WIDTH = 10;
const CELL_HEIGHT = 5;
const HGAP = 15;
const VGAP = 5;

function SvgArea({ goRects }) {
  return (
    <svg width="500" height="500">
      {goRects.map(({ id, x, y, height }) =>
        <rect key={id} style={{fill: 'grey'}}
          x={x*(CELL_WIDTH+HGAP)} y={y*(CELL_HEIGHT+VGAP)}
          width={CELL_WIDTH} height={height*(CELL_HEIGHT+VGAP) + CELL_HEIGHT} />)}
    </svg>
  );
}


function App() {
  const [sql, setSql] = useState(null);
  const [db, setDb] = useState(null);
  const [goRects, setGoRects] = useState(null);

  useEffect(() => {
    const helper = async () => {
      const sql0 = await initSqlJs({
        locateFile: filename => `/${filename}`
      });
      setSql(sql0);
    }
    helper();
  }, [sql]);

  useEffect(() => {
    if (!db) { return; }

    let timeEventsPopulated = false;
    while (!timeEventsPopulated) {
      const [{values: [[countBefore]]}] = db.exec(`SELECT COUNT(*) FROM time_events`);
      db.exec(`INSERT INTO time_events SELECT * FROM new_spawn_child_events`);
      db.exec(`INSERT INTO time_events SELECT * FROM new_goroutine_start_events`);
      const [{values: [[countAfter]]}] = db.exec(`SELECT COUNT(*) FROM time_events`);
      timeEventsPopulated = (countBefore === countAfter);
    }
    const rects = db.exec(`SELECT id, x, y, height FROM goroutine_rects`)[0].values
      .map(([id, x, y, height]) => ({ id, x, y, height }));
    setGoRects(rects);
  }, [db]);

  const onFileSelect = (e) => {
    const f = e.target.files[0];
    const r = new FileReader();
    r.onload = function () {
      const bytes = new Uint8Array(r.result);
      setDb(new sql.Database(bytes));
    }
    r.readAsArrayBuffer(f);
  }

  const dbIsLoaded = !!goRects;
  // hide input, if db is created
  const fileInputDisplay = dbIsLoaded ? 'none' : 'block';

  return (
    <div className="App">
      <header style={{display: fileInputDisplay}} className="App-header">
        <input type="file" onChange={onFileSelect} />
      </header>
      {dbIsLoaded && <SvgArea goRects={goRects} />}
    </div>
  );
}

export default App;
