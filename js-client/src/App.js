import React, { useState, useCallback } from 'react';
import Select from "react-select";

import RunPlot from './RunPlot';


function UpdateTraces(data, selectedOption) {
  var traces = []
  console.log(selectedOption)
  //for (let line of Object.entries(data.linesByTimestamp["120000000000"].lines)) {
  for (let line of Object.entries(data.linesByTimestamp[selectedOption].lines)) {
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
  return traces
}

/**
 * Main application component
 *
 * @returns
 */
 export const App = () => {
  const [timesteps, setTimesteps] = useState()
  const [selectedOption, setSelectedOption] = useState()

  const [traces, setTraces] = useState(function() {
    fetch("http://localhost:8080/encode").then(function(response) {
      return response.json();
    }).then(function(data) {
      var timesteps = []
      for (let step of Object.entries(data.linesByTimestamp)) {
        timesteps.push({value: step[0], label: step[0]})
      }
      setTimesteps(timesteps)
      setTraces(data)
      //console.log(JSON.stringify(data))
      setSelectedOption(timesteps[0])
      //console.log(JSON.stringify(timesteps[0]))

    }).catch(error => console.log(error.message))
  });

  const onOptionChange = useCallback((option) => setSelectedOption(option), []);

  return (
    <div>
      <Select
        value={selectedOption}
        onChange={onOptionChange}
        options={timesteps}
      />
      {(traces && selectedOption) ? <RunPlot traces={traces && traces.linesByTimestamp[selectedOption.value]} /> : ""}
    </div>
  );
}
export default App;
