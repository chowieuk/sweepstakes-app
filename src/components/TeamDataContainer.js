import React from "react";

import TeamDataBox from "./TeamDataBox";

export default function TeamDataContainer(props) {
  const { flag, name_en, name_fa, team_id, ...currentStats } = props.stats;

  return (
    <div className="teamdata-container">
      <TeamDataBox header="MP" value={currentStats.mp} />
      <TeamDataBox header="W" value={currentStats.w} />
      <TeamDataBox header="L" value={currentStats.l} />
      <TeamDataBox header="GF" value={currentStats.gf} />
      <TeamDataBox header="GA" value={currentStats.ga} />
      <TeamDataBox header="GD" value={currentStats.gd} />
      <TeamDataBox header="Pts" value={currentStats.pts} />
    </div>
  );
}
