import React from 'react';
import Plot from 'react-plotly.js'

function UpdateTraces(data) {
  var traces = []
  if (data) {
    for (let line of Object.entries(data.lines)) {
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
  }
  return traces
}

/**
 * Main application component
 *
 * @returns
 */
 export const RunPlot = (props) => {
  return (
    <div>
      <Plot
        data={UpdateTraces(props.traces)}
        layout={{
          width: 900,
          height: 800,
          title: `Simple 3D Scatter`
        }}
      />
    </div>
  );
}
export default RunPlot;
