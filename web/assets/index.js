import { Terminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import '@xterm/xterm/css/xterm.css';
import "nes.css/css/nes.min.css";
import './layout.css';
import './style.css';

function active(e) {
    const currentActive = document.querySelector(".nes-container.with-tabs .tab.active");
    currentActive?.classList.remove("active");
    e.classList.add("active");
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

        // Create and load the fit addon
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

            // Force a reflow to get accurate dimensions
            container.style.display = 'block';
            container.offsetHeight; // Force reflow

            // Get the actual container dimensions after reflow
            const containerRect = container.getBoundingClientRect();
            const availableWidth = Math.max(containerRect.width - 20, 800); // Account for padding, minimum width
            const availableHeight = Math.max(containerRect.height - 20, 400); // Account for padding, minimum height

            console.log(`Container dimensions: ${availableWidth}x${availableHeight}`);

            // Use more accurate character dimensions based on xterm's defaults
            const charWidth = 8.4; // More accurate for monospace fonts
            const lineHeight = 17; // Standard line height for terminals

            // Calculate terminal dimensions with proper minimums
            const cols = Math.max(120, Math.floor(availableWidth / charWidth));
            const rows = Math.max(33, Math.floor(availableHeight / lineHeight));

            console.log(`Calculated terminal size: ${cols}x${rows}`);

            // Only resize if dimensions have changed significantly
            if (Math.abs(term.cols - cols) > 1 || Math.abs(term.rows - rows) > 1) {
                console.log(`Resizing terminal from ${term.cols}x${term.rows} to ${cols}x${rows}`);
                term.resize(cols, rows);

                // Fit the terminal to the exact container size
                if (fitAddon) {
                    setTimeout(() => fitAddon.fit(), 100);
                }

                // Notify backend of the new size
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

    // if the url doesn't have a # fragment, add #rules
    let url = window.location.href;
    if (url.indexOf('#') === -1) {
        url += '#rules';
        window.location.href = url;
        let e = document.querySelector(".nes-container.with-tabs .tabs .tab:first-child");
        if (e) active(e);
    } else {
        const hash = url.substring(url.indexOf('#'));
        let e = document.querySelector(`.nes-container.with-tabs .tabs .tab a[href="${hash}"]`);
        if (e) active(e.parentElement);
        if (hash === '#term') {
            initializeTerminal();
        }
    }

    let tabs = document.querySelectorAll(".nes-container.with-tabs > .tabs > .tab");
    for (let i = 0; i < tabs.length; i++) {
        tabs[i].onclick = function () {
            active(tabs[i]);

            const link = tabs[i].querySelector('a');
            if (link && link.getAttribute('href') === '#term') {
                initializeTerminal();
            }
        };
    }
}
