import React, { useState, useCallback } from 'react';
import Select from "react-select";

import RunPlot from './RunPlot';

function microToString(timestep) {
  var nano = 1000 * 1000 * 1000
  var minute = 60 * nano
  var minutes = Math.floor(timestep / minute);

  var seconds = ((timestep % minute) / nano).toFixed(0);
  return minutes + "m:" + (seconds < 10 ? '0' : '') + seconds + "s";
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
        timesteps.push({value: step[0], label: microToString(step[0])})
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
