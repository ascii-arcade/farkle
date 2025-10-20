import { Terminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import '@xterm/xterm/css/xterm.css';
import "nes.css/css/nes.min.css";
import './layout.css';
import './style.css';

function saveActiveTab(tabHash, expirationHours = 24) {
    const expirationTime = Date.now() + (expirationHours * 60 * 60 * 1000);
    const tabData = {
        tab: tabHash,
        expires: expirationTime
    };
    localStorage.setItem('activeTab', JSON.stringify(tabData));
}

function getActiveTab() {
    try {
        const saved = localStorage.getItem('activeTab');
        if (!saved) return null;

        const tabData = JSON.parse(saved);

        if (Date.now() > tabData.expires) {
            localStorage.removeItem('activeTab');
            return null;
        }

        return tabData.tab;
    } catch (error) {
        localStorage.removeItem('activeTab');
        return null;
    }
}

function active(e) {
    const currentActive = document.querySelector(".nes-container.with-tabs .tab.active");
    currentActive?.classList.remove("active");
    e.classList.add("active");

    const currentContent = document.querySelector(".nes-container.with-tabs .content.active");
    currentContent?.classList.remove("active");

    const link = e.querySelector('a');
    if (link) {
        const href = link.getAttribute('href');
        if (href && href.startsWith('#')) {
            const targetContent = document.querySelector(href);
            if (targetContent) {
                targetContent.classList.add('active');
            }

            saveActiveTab(href, 1);

            const scrollTop = window.pageYOffset || document.documentElement.scrollTop;
            const scrollLeft = window.pageXOffset || document.documentElement.scrollLeft;
            history.replaceState(null, null, href);
            window.scrollTo(scrollLeft, scrollTop);
        }
    }
}

window.onload = function () {
    let terminalInitialized = false;
    let term = null;
    let fitAddon = null;
    let ws = null;
    let currentInput = '';
    let sshConnected = false;

    function initializeTerminal() {
        if (terminalInitialized) return;

        console.log('Initializing terminal...');
        terminalInitialized = true;

        term = new Terminal({
            cols: 120,
            rows: 33,
            fontSize: 12,
            fontFamily: 'Monaco, Menlo, "Ubuntu Mono", monospace',
            theme: {
                background: '#212529',
                foreground: '#ffffff'
            },
            cursorBlink: true
        });

        fitAddon = new FitAddon();
        term.loadAddon(fitAddon);

        term.open(document.getElementById('xterm-container'));

        function fitTerminal() {
            console.log('Fitting terminal to container size...');

            const container = document.getElementById('xterm-container');
            if (!container || !term) {
                console.log('Container or terminal not available for fitting');
                return;
            }

            container.style.display = 'block';
            container.offsetHeight;

            const containerRect = container.getBoundingClientRect();
            const availableWidth = Math.max(containerRect.width - 20, 800); // Account for padding, minimum width
            const availableHeight = Math.max(containerRect.height - 20, 400); // Account for padding, minimum height

            console.log(`Container dimensions: ${availableWidth}x${availableHeight}`);

            const charWidth = 8.4; // More accurate for monospace fonts
            const lineHeight = 17; // Standard line height for terminals

            const cols = Math.max(120, Math.floor(availableWidth / charWidth));
            const rows = Math.max(33, Math.floor(availableHeight / lineHeight));

            console.log(`Calculated terminal size: ${cols}x${rows}`);

            if (Math.abs(term.cols - cols) > 1 || Math.abs(term.rows - rows) > 1) {
                console.log(`Resizing terminal from ${term.cols}x${term.rows} to ${cols}x${rows}`);
                term.resize(cols, rows);

                if (fitAddon) {
                    setTimeout(() => fitAddon.fit(), 100);
                }

                if (socket && socket.readyState === WebSocket.OPEN) {
                    const resizeMessage = JSON.stringify({
                        type: 'resize',
                        cols: cols,
                        rows: rows
                    });
                    console.log('Sending resize message to backend:', resizeMessage);
                    socket.send(resizeMessage);
                }
            }
        }

        setTimeout(fitTerminal, 100);

        window.addEventListener('resize', fitTerminal);

        const wsProtocol = process.env.WS_PROTOCOL;
        console.log('Creating WebSocket connection to:', `${wsProtocol}://${window.location.host}/ws`);
        ws = new WebSocket(`${wsProtocol}://${window.location.host}/ws`);

        ws.onopen = function () {
            console.log('WebSocket connected');
            term.write('Connecting to SSH server...\r\n');

            const cols = term.cols;
            const rows = term.rows;
            ws.send(`RESIZE:${cols},${rows}`);
            console.log(`Sent initial size: ${cols}x${rows}`);
        };

        ws.onmessage = function (event) {
            sshConnected = true;

            if (event.data instanceof Blob) {
                const reader = new FileReader();
                reader.onload = function () {
                    const arrayBuffer = reader.result;
                    const uint8Array = new Uint8Array(arrayBuffer);
                    term.write(uint8Array);
                };
                reader.readAsArrayBuffer(event.data);
            } else {
                term.write(event.data);
            }
        };

        ws.onclose = function () {
            console.log('WebSocket disconnected');
            if (term) {
                term.write('\r\nConnection closed\r\n');
            }
        };

        ws.onerror = function (error) {
            console.error('WebSocket error:', error);
            if (term) {
                term.write('\r\nConnection error\r\n');
            }
        };

        term.onData(data => {
            if (sshConnected) {
                if (ws && ws.readyState === WebSocket.OPEN) {
                    ws.send(data);
                }
            } else {
                const code = data.charCodeAt(0);

                if (code === 13) { // Enter key
                    if (ws && ws.readyState === WebSocket.OPEN) {
                        ws.send(currentInput + '\n');
                    }
                    term.write('\r\n');
                    currentInput = '';
                } else if (code === 127 || code === 8) { // Backspace or Delete
                    if (currentInput.length > 0) {
                        currentInput = currentInput.slice(0, -1);
                        term.write('\b \b');
                    }
                } else if (code >= 32) { // Printable characters
                    currentInput += data;
                    term.write(data);
                }
            }
        });
    }

    let url = window.location.href;
    let targetHash = null;

    if (url.indexOf('#') !== -1) {
        targetHash = url.substring(url.indexOf('#'));
    } else {
        const savedTab = getActiveTab();
        if (savedTab) {
            targetHash = savedTab;
        }
    }

    if (targetHash) {
        let e = document.querySelector(`.nes-container.with-tabs .tabs .tab a[href="${targetHash}"]`);
        if (e) {
            active(e.parentElement);
            if (targetHash === '#term') {
                initializeTerminal();
            }
        } else {
            let firstTab = document.querySelector(".nes-container.with-tabs .tabs .tab:first-child");
            if (firstTab) active(firstTab);
        }
    } else {
        let firstTab = document.querySelector(".nes-container.with-tabs .tabs .tab:first-child");
        if (firstTab) active(firstTab);
    }

    let tabs = document.querySelectorAll(".nes-container.with-tabs > .tabs > .tab");
    for (let i = 0; i < tabs.length; i++) {
        tabs[i].onclick = function (event) {
            event.preventDefault();

            active(tabs[i]);

            const link = tabs[i].querySelector('a');
            if (link && link.getAttribute('href') === '#term') {
                initializeTerminal();
            }
        };
    }
}
