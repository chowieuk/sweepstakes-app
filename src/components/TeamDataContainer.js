import React from "react";

import TeamDataBox from "./TeamDataBox";

export default function TeamDataContainer(props) {

const {flag, name_en, name_fa, team_id, ...currentStats} = props.stats;

  return <>

           <TeamDataBox
            className="teamdata-container"
            header= "MP"
            value={currentStats.mp}
          />
           <TeamDataBox
            className="teamdata-container"
            header= "W"
            value={currentStats.w}
          />
           <TeamDataBox
            className="teamdata-container"
            header= "D"
            value={currentStats.d}
          />
           <TeamDataBox
            className="teamdata-container"
            header= "L"
            value={currentStats.l}
          />
           <TeamDataBox
            className="teamdata-container"
            header= "GF"
            value={currentStats.gf}
          />
           <TeamDataBox
            className="teamdata-container"
            header= "GA"
            value={currentStats.ga}
          />
           <TeamDataBox
            className="teamdata-container"
            header= "GD"
            value={currentStats.gd}
          />
           <TeamDataBox
            className="teamdata-container"
            header= "Pts"
            value={currentStats.pts}
          />

      </>;
}
