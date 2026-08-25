/* buck-ccm 前端：页面交互与结果渲染。所有数值来自后端 API，不做求解。 */

(function () {
  "use strict";

  var fields = ["vin", "d", "l", "c", "ts", "r"];

  var els = {
    loadExample: document.getElementById("load-example"),
    compute: document.getElementById("compute"),
    result: document.getElementById("result"),
    error: document.getElementById("error"),
    hint: document.getElementById("hint"),
    modeBadge: document.getElementById("mode-badge"),
    vout: document.getElementById("vout"),
    k: document.getElementById("k"),
    kcrit: document.getElementById("kcrit"),
    deltaIL: document.getElementById("delta-il"),
    deltaVC: document.getElementById("delta-vc"),
    region: document.getElementById("region"),
    wave: document.getElementById("wave")
  };

  function getSpec() {
    var spec = {};
    fields.forEach(function (name) {
      var raw = document.getElementById(name).value.trim();
      var v = parseFloat(raw);
      if (raw === "" || isNaN(v)) {
        throw new Error("参数 " + name + " 必须是数值，当前输入：" + raw);
      }
      spec[name] = v;
    });
    return spec;
  }

  function fillForm(spec) {
    fields.forEach(function (name) {
      document.getElementById(name).value = spec[name];
    });
  }

  function showError(message) {
    els.result.classList.add("hidden");
    els.error.textContent = "后端返回：" + message;
    els.error.classList.remove("hidden");
    els.hint.classList.add("hidden");
  }

  function fmt(v, digits) {
    if (typeof v !== "number" || !isFinite(v)) return String(v);
    if (Math.abs(v) >= 1000 || (v !== 0 && Math.abs(v) < 0.001)) {
      return v.toExponential(4);
    }
    return v.toFixed(digits === undefined ? 4 : digits);
  }

  // drawWave 用后端返回的点列画电感电流三角波，禁止自行构造三角形。
  function drawWave(points, period) {
    var svg = els.wave;
    svg.innerHTML = "";
    if (!points || points.length < 2) {
      svg.innerHTML = '<text x="12" y="20" class="note">无波形数据</text>';
      return;
    }
    var W = 720, H = 240, pad = 28;
    var minT = points[0].t, maxT = period;
    var minI = Infinity, maxI = -Infinity;
    points.forEach(function (p) {
      if (p.i < minI) minI = p.i;
      if (p.i > maxI) maxI = p.i;
    });
    if (minI === maxI) { minI -= 1; maxI += 1; }
    var spanI = maxI - minI;

    function x(t) { return pad + ((t - minT) / (maxT - minT)) * (W - 2 * pad); }
    function y(i) { return H - pad - ((i - minI) / spanI) * (H - 2 * pad); }

    var d = "";
    points.forEach(function (p, idx) {
      d += (idx === 0 ? "M" : "L") + x(p.t).toFixed(2) + " " + y(p.i).toFixed(2);
    });

    var grid = document.createElementNS("http://www.w3.org/2000/svg", "line");
    grid.setAttribute("x1", pad);
    grid.setAttribute("y1", y(0));
    grid.setAttribute("x2", W - pad);
    grid.setAttribute("y2", y(0));
    grid.setAttribute("class", "grid");
    svg.appendChild(grid);

    var poly = document.createElementNS("http://www.w3.org/2000/svg", "path");
    poly.setAttribute("d", d);
    poly.setAttribute("class", "current");
    svg.appendChild(poly);

    var label = document.createElementNS("http://www.w3.org/2000/svg", "text");
    label.setAttribute("x", pad + 4);
    label.setAttribute("y", 16);
    label.setAttribute("class", "note");
    label.textContent = "i_L(t)：min " + fmt(minI) + " A，max " + fmt(maxI) + " A";
    svg.appendChild(label);
  }

  // compute 同时请求模式与纹波两个接口。
  function compute() {
    var spec;
    try {
      spec = getSpec();
    } catch (err) {
      showError(err.message);
      return;
    }
    var payload = JSON.stringify(spec);
    els.hint.classList.add("hidden");

    Promise.all([
      postJSON("/api/mode", payload),
      postJSON("/api/ripple", payload)
    ]).then(function (replies) {
      var mode = replies[0];
      var ripple = replies[1];
      els.modeBadge.textContent = mode.mode;
      els.modeBadge.className = "mode-badge " + mode.mode.toLowerCase();
      els.vout.textContent = fmt(mode.vout) + " V";
      els.k.textContent = fmt(mode.k);
      els.kcrit.textContent = fmt(mode.kcrit);
      els.deltaIL.textContent = fmt(ripple.delta_il) + " A";
      els.deltaVC.textContent = fmt(ripple.delta_vc) + " V";
      els.region.textContent = mode.region;
      els.error.classList.add("hidden");
      els.result.classList.remove("hidden");
      drawWave(ripple.points, ripple.period);
    }).catch(function (err) {
      showError(err.message);
    });
  }

  // postJSON 发 POST 并把非 2xx 响应转成后端错误文案。
  function postJSON(url, payload) {
    return fetch(url, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: payload
    }).then(function (resp) {
      return resp.json().catch(function () {
        throw new Error("后端响应不是 JSON（HTTP " + resp.status + "）");
      }).then(function (body) {
        if (!resp.ok) {
          throw new Error(body.error || ("HTTP " + resp.status));
        }
        return body;
      });
    });
  }

  function loadExample() {
    fetch("/example/12v-5v.json")
      .then(function (resp) {
        if (!resp.ok) throw new Error("加载算例失败（HTTP " + resp.status + "）");
        return resp.json();
      })
      .then(function (spec) {
        fillForm(spec);
        els.hint.textContent = "已加载 example/12v-5v.json，可点「计算」。";
        els.hint.classList.remove("hidden");
      })
      .catch(function (err) {
        showError(err.message);
      });
  }

  els.compute.addEventListener("click", compute);
  els.loadExample.addEventListener("click", loadExample);
})();
