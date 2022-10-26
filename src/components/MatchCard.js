import React from "react";

export default function MatchCard({ props }) {
  return (
    <div>
      <table class="matchCard">
        <tbody>
          <tr>
            <td colspan="6">
              <div class="cardText">
                <span>Group A · Matchday 1 of 3</span>
              </div>
            </td>
          </tr>
          <tr>
            <td colspan="3"></td>
            <td rowspan="5"></td>
            <td class="textDivider" rowspan="5">
              <div>
                <div>
                  <div>
                    <div class="cardText">Nov 20</div>
                    <div class="cardText">21:45</div>
                  </div>
                </div>
              </div>
            </td>
            <td></td>
          </tr>
          <tr>
            <td></td>
          </tr>
          <tr>
            <td>
              <img
                alt=""
                src="//ssl.gstatic.com/onebox/media/sports/logos/h0FNA5YxLzWChHS5K0o4gw_48x48.png"
                style="width: 24px; height: 24px"
              />
            </td>
            <td class="cardText">
              <div data-df-team-mid="/m/046zk0">
                <span>Qatar</span>
                <span></span>
              </div>
            </td>
            <td></td>
          </tr>
          <tr>
            <td data-df-team-mid="/m/03zj_3">
              <img
                alt=""
                src="//ssl.gstatic.com/onebox/media/sports/logos/AKqvkBpIyr-iLOK7Ig7-yQ_48x48.png"
                style="width: 24px; height: 24px"
              />
            </td>
            <td class="cardText">
              <div class="ellipsisize" data-df-team-mid="/m/03zj_3">
                <span>Ecuador</span>
                <span></span>
              </div>
            </td>
            <td></td>
          </tr>
          <tr>
            <td></td>
          </tr>
          <tr></tr>
        </tbody>
      </table>
    </div>
  );
}
