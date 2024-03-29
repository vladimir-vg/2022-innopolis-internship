import { useEffect, useState } from 'react';
import CodeMirror from '@uiw/react-codemirror';
import { StreamLanguage } from '@codemirror/language';
import { go } from '@codemirror/legacy-modes/mode/go';

import './App.css';
import initSqlJs from 'sql.js';

const CELL_WIDTH = 11;
const CELL_HEIGHT = 11;
const HGAP = 15;
const VGAP = 3;
const SPAWN_LINE_WIDTH = 1;
const SPAWN_OUTLINE_WIDTH = 6;

function GoroutineBody({ id, x, y, height, selectedId, setSelectedId }) {
  // toggle selection
  const onClick = (e) => setSelectedId((id === selectedId) ? null : id);
  const className = 'GoroutineBody' + (id === selectedId ? ' selected' : '');

  return <g className={className}>
    <rect className="GoroutineBody-main"
      x={x*(CELL_WIDTH+HGAP)} y={y*(CELL_HEIGHT+VGAP)}
      width={CELL_WIDTH} height={height*(CELL_HEIGHT+VGAP) + CELL_HEIGHT} />
    <rect className="GoroutineBody-header" onClick={onClick}
      x={x*(CELL_WIDTH+HGAP)} y={y*(CELL_HEIGHT+VGAP)}
      width={CELL_WIDTH} height={CELL_HEIGHT} />
  </g>;
}

function SpawnLine({ id, x1, y1, x2, y2, selectedId, setSelectedId }) {
  const onClick = (e) => setSelectedId((id === selectedId) ? null : id);
  const className = 'SpawnLine' + (id === selectedId ? ' selected' : '');
  return <g className={className} onClick={onClick}>
    <circle className="SpawnLine-parentPoint"
      cx={x1*(CELL_WIDTH+HGAP) + CELL_WIDTH/2}
      cy={y1*(CELL_HEIGHT+VGAP) + CELL_HEIGHT/2}
      r={3} />
    <rect className="SpawnLine-outline"
      x={x1*(CELL_WIDTH+HGAP) + CELL_WIDTH}
      y={y1*(CELL_HEIGHT+VGAP) + CELL_HEIGHT/2 - SPAWN_OUTLINE_WIDTH/2}
      width={(x2-x1)*(CELL_WIDTH+HGAP) - CELL_WIDTH/2}
      height={SPAWN_OUTLINE_WIDTH} />
    <rect className="SpawnLine-outline"
      x={x2*(CELL_WIDTH+HGAP) + CELL_WIDTH/2 - SPAWN_OUTLINE_WIDTH/2}
      y={y1*(CELL_HEIGHT+VGAP) + CELL_HEIGHT/2 - SPAWN_OUTLINE_WIDTH/2}
      width={SPAWN_OUTLINE_WIDTH}
      height={(y2-y1)*(CELL_HEIGHT+VGAP) - SPAWN_OUTLINE_WIDTH/2} />

    <line className="SpawnLine-line" style={{strokeWidth: SPAWN_LINE_WIDTH}}
      x1={x1*(CELL_WIDTH+HGAP) + CELL_WIDTH/2}
      y1={y1*(CELL_HEIGHT+VGAP) + CELL_HEIGHT/2}
      x2={x2*(CELL_WIDTH+HGAP) + CELL_WIDTH/2}
      y2={y1*(CELL_HEIGHT+VGAP) + CELL_HEIGHT/2} />
    <line className="SpawnLine-line" style={{strokeWidth: SPAWN_LINE_WIDTH}}
      x1={x2*(CELL_WIDTH+HGAP) + CELL_WIDTH/2}
      y1={y1*(CELL_HEIGHT+VGAP) + CELL_HEIGHT/2}
      x2={x2*(CELL_WIDTH+HGAP) + CELL_WIDTH/2}
      y2={y2*(CELL_HEIGHT+VGAP)} />
  </g>;
}

function SvgArea({ figures, selectedId, setSelectedId }) {
  const { rects, spawnLines } = figures;
  return (
    <svg width="500" height="500">
      {rects.map(({ id, x, y, height }) =>
        <GoroutineBody key={id} id={id}
          selectedId={selectedId} setSelectedId={setSelectedId}
          x={x} y={y} height={height} />)}
      {spawnLines.map(({ id, x1, y1, x2, y2 }, i) =>
        <SpawnLine key={id} id={id}
          selectedId={selectedId} setSelectedId={setSelectedId}
          x1={x1} y1={y1} x2={x2} y2={y2} />)}
    </svg>
  );
}



function App() {
  const [sql, setSql] = useState(null);
  const [db, setDb] = useState(null);
  const [figures, setFigures] = useState(null);

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
    const rects = db.exec(`SELECT id, x, y, height, filename, line FROM goroutine_rects`)[0].values
      .map(([id, x, y, height, filename, line]) => ({ id, x, y, height, filename, line }));
    const spawnLines = db.exec(`SELECT id, parentId, childId, x1, y1, x2, y2, filename, line FROM spawn_lines`)[0].values
      .map(([id, parentId, childId, x1, y1, x2, y2, filename, line]) =>
        ({ id, parentId, childId, x1, y1, x2, y2, filename, line }));
    setFigures({ rects, spawnLines });
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

  const [selectedId, setSelectedId] = useState(null);
  const [sourceContent, setSourceContent] = useState('');

  useEffect(() => {
    if (!db) { return; }
    if (selectedId == null) {
      setSourceContent('');
      return;
    }

    for (const rect of [...figures.rects, figures.spawnLines]) {
      const { id, filename, line } = rect;
      if (id === selectedId) {
        const [content] = db.exec(`SELECT content FROM files WHERE filename = ?`, [filename])[0].values[0];
        const contentStr = new TextDecoder().decode(content);
        setSourceContent(contentStr);
        return;
      }
    }
  }, [selectedId, db]);

  const dbIsLoaded = !!figures;
  // hide input, if db is created
  const fileInputDisplay = dbIsLoaded ? 'none' : 'block';

  return (
    <div className="App">
      <header style={{display: fileInputDisplay}} className="App-header">
        <input type="file" onChange={onFileSelect} />
      </header>
      <section>
        {dbIsLoaded && <SvgArea figures={figures} selectedId={selectedId} setSelectedId={setSelectedId} />}
      </section>
      <section>
        {dbIsLoaded &&
          <CodeMirror
            extensions={[StreamLanguage.define(go)]}
            value={sourceContent}/>}
      </section>
    </div>
  );
}

export default App;
