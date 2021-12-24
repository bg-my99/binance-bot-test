import React, { useState, useCallback } from 'react';
import Plot from 'react-plotly.js'
import Select from "react-select";


/**
 * Main application component
 *
 * @returns
 */
function App() {
  const [traces, setTraces] = useState(function() {
    fetch("http://localhost:8080/encode").then(function(response) {
      return response.json();
    }).then(function(data) {
      var traces = []
      for (let line of Object.entries(data.linesByTimestamp["120000000000"].lines)) {
        var x = []
        var y = []
        var z = []
        for (let point of Object.entries(line[1].points.sort((a, b) => (a.rsiSell > b.rsiSell) ? 1 : -1))) {
          x.push(line[0])
          y.push(point[1].rsiSell)
          z.push(point[1].pnl)
        }
        traces.push({
          x: x,
          y: y,
          z: z,
          type: 'scatter3d',
          mode: 'lines'  
        })
    }
      setTraces(traces)
      //console.log(JSON.stringify(data))    
    }).catch(error => console.log(error.message))
  });
  const [selectedOption, setSelectedOption] = useState()

  const icecreams = [
    { value: 'chocolate', label: 'Chocolate' },
    { value: 'strawberry', label: 'Strawberry' },
    { value: 'vanilla', label: 'Vanilla' },
  ];

  const onOptionChange = useCallback((option) => setSelectedOption(option), []);

  return (
    <div>
      <Select
        value={selectedOption}
        onChange={onOptionChange}
        options={icecreams}
      />
      <Plot
        data={traces}
        layout={{
          width: 900,
          height: 800,
          title: `Simple 3D Scatter`
        }}
      />
    </div>
  );
}
export default App;
