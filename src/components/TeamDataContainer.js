import React from "react";

import TeamDataBox from "./TeamDataBox";

export default function TeamDataContainer(props) {

const {flag, name_en, name_fa, team_id, ...currentStats} = props.stats;

console.log(currentStats)
        // for (let i = 1; i < 9; i++) {
        // console.log(props.stats[1])

        // } 
  return <>



          {props.stats.mp}
      </>;
}
