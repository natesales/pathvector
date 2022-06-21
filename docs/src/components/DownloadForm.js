import Admonition from '@theme/Admonition';
import React from "react";

function toggleActive(e) {
    if (e.target.classList.contains("active")) {
        e.target.classList.remove("active")
    } else {
        e.target.classList.add("active")
    }
}

export const DownloadForm = () => (
    <>
        <Admonition type="note" title="Download">
            <div className="download-form">
                <div className="download-col">
                    <label>OS: </label>
                    <select id="os">
                        <option value="linux">Linux</option>
                        <option value="windows">Windows</option>
                        <option value="macos">macOS</option>
                        <option value="freebsd">FreeBSD</option>
                    </select>
                </div>

                <div className="download-col">
                    <label>Architecture: </label>
                    <select id="architecture">
                        <option value="amd64">amd64</option>
                        <option value="mips">mips</option>
                    </select>
                </div>

                <div className="download-col">
                    <label>Platform: </label>
                    <select id="platform">
                        <option value="debian">Debian</option>
                        <option value="centos">CentOS</option>
                        <option value="arista">Arista</option>
                        <option value="juniper">Juniper</option>
                        <option value="cisco">Cisco</option>
                        <option value="mikrotik">Mikrotik</option>
                        <option value="binary">Binary</option>
                    </select>
                </div>

                <div className="download-col">
                    <button onClick={() => {
                        let os = document.getElementById("os").value;
                        let architecture = document.getElementById("architecture").value;
                        let platform = document.getElementById("platform").value;
                    }}>Download
                    </button>
                </div>
            </div>
        </Admonition>

        <p>Plugins:</p>

        <button onClick={(e) => toggleActive(e)} className="plugin-select">repo.pathvector.io/enterprise</button>
        <button onClick={(e) => toggleActive(e)} className="plugin-select">repo.pathvector.io/enterprise</button>
        <button onClick={(e) => toggleActive(e)} className="plugin-select">repo.pathvector.io/enterprise</button>
        <button onClick={(e) => toggleActive(e)} className="plugin-select">repo.pathvector.io/enterprise</button>
        <button onClick={(e) => toggleActive(e)} className="plugin-select">repo.pathvector.io/enterprise</button>
    </>
)
