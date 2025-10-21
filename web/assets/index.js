import { Terminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import '@xterm/xterm/css/xterm.css';
import "nes.css/css/nes.min.css";
import './layout.css';
import './style.css';

let terminalInitialized = false;
let term = null;
let fitAddon = null;
let ws = null;
let currentInput = '';
let sshConnected = false;
let narrowWindowRetryCount = 0;
const MAX_NARROW_WINDOW_RETRIES = 5;
let clientSessionId = null;

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

function saveClientSessionId(sessionId, expirationHours = 24) {
    const expirationTime = Date.now() + (expirationHours * 60 * 60 * 1000);
    const sessionData = {
        sessionId: sessionId,
        expires: expirationTime,
        created: Date.now()
    };
    localStorage.setItem('sshClientSession', JSON.stringify(sessionData));
    console.log('SSH session cached until:', new Date(expirationTime).toLocaleString());
}

function getClientSessionId() {
    try {
        const saved = localStorage.getItem('sshClientSession');
        if (!saved) return null;

        const sessionData = JSON.parse(saved);

        if (Date.now() > sessionData.expires) {
            localStorage.removeItem('sshClientSession');
            console.log('SSH session expired, will generate new key pair');
            return null;
        }

        const hoursLeft = Math.round((sessionData.expires - Date.now()) / (1000 * 60 * 60));
        console.log(`Reusing SSH session (${hoursLeft}h remaining):`, sessionData.sessionId);
        return sessionData.sessionId;
    } catch (error) {
        localStorage.removeItem('sshClientSession');
        console.log('Error retrieving SSH session, will generate new key pair');
        return null;
    }
}

function generateClientSessionId() {
    const array = new Uint8Array(16);
    crypto.getRandomValues(array);
    return Array.from(array, byte => byte.toString(16).padStart(2, '0')).join('');
}

function updateSSHKeyStatus(message, isNew = false) {
    const statusElement = document.getElementById('ssh-key-status');
    if (statusElement) {
        statusElement.textContent = `🔑 ${message}`;
        statusElement.className = isNew ? 'nes-text is-warning' : 'nes-text is-success';
        statusElement.hidden = false;
        
        if (message.includes('Generated') || message.includes('cached')) {
            setTimeout(() => {
                if (statusElement.textContent.includes(message.split(':')[0])) {
                    statusElement.hidden = true;
                }
            }, 5000);
        }
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

    clientSessionId = getClientSessionId();
    if (!clientSessionId) {
        clientSessionId = generateClientSessionId();
        saveClientSessionId(clientSessionId);
        console.log('Generated new SSH session ID:', clientSessionId);
        updateSSHKeyStatus('Generated new SSH key (24h cache)', true);
    } else {
        const saved = localStorage.getItem('sshClientSession');
        if (saved) {
            const sessionData = JSON.parse(saved);
            const hoursLeft = Math.round((sessionData.expires - Date.now()) / (1000 * 60 * 60));
            updateSSHKeyStatus(`Using cached SSH key (${hoursLeft}h remaining)`, false);
        }
    }

    const wsProtocol = process.env.WS_PROTOCOL;
    console.log('Creating WebSocket connection to:', `${wsProtocol}://${window.location.host}/ws`);
    ws = new WebSocket(`${wsProtocol}://${window.location.host}/ws`);

    ws.onopen = function () {
        console.log('WebSocket connected');
        term.write('Connecting to SSH server...\r\n');

        const sessionMessage = JSON.stringify({
            type: 'session',
            sessionId: clientSessionId
        });
        ws.send(sessionMessage);
        console.log('Sent session ID for SSH key caching');

        const cols = term.cols;
        const rows = term.rows;
        ws.send(`RESIZE:${cols},${rows}`);
        console.log(`Sent initial size: ${cols}x${rows}`);
    };

    ws.onmessage = function (event) {
        sshConnected = true;

        let data;
        let dataString = '';

        if (event.data instanceof Blob) {
            const reader = new FileReader();
            reader.onload = function () {
                const arrayBuffer = reader.result;
                const uint8Array = new Uint8Array(arrayBuffer);
                data = uint8Array;
                
                // Convert to string to check for "Window too narrow" message
                dataString = new TextDecoder().decode(uint8Array);
                
                handleTerminalData(data, dataString);
            };
            reader.readAsArrayBuffer(event.data);
        } else {
            data = event.data;
            dataString = event.data;
            handleTerminalData(data, dataString);
        }
    };

    function handleTerminalData(data, dataString) {
        // Check for various "Window too narrow" or size-related error messages
        const narrowPatterns = [
            'window too narrow',
            'terminal too narrow', 
            'screen too narrow',
            'terminal too small',
            'window too small',
            'insufficient screen width',
            'insufficient terminal width'
        ];
        
        const containsNarrowError = narrowPatterns.some(pattern => 
            dataString.toLowerCase().includes(pattern)
        );
        
        if (containsNarrowError) {
            console.log('Detected narrow window error message:', dataString.trim());
            
            if (narrowWindowRetryCount < MAX_NARROW_WINDOW_RETRIES) {
                narrowWindowRetryCount++;
                
                // Increase terminal size more aggressively each retry
                const extraCols = narrowWindowRetryCount * 15;
                const extraRows = narrowWindowRetryCount * 8;
                const newCols = Math.max(term.cols + extraCols, 150);
                const newRows = Math.max(term.rows + extraRows, 45);
                
                console.log(`Auto-resize attempt ${narrowWindowRetryCount}/${MAX_NARROW_WINDOW_RETRIES}: ${term.cols}x${term.rows} -> ${newCols}x${newRows}`);
                
                term.resize(newCols, newRows);
                
                if (ws && ws.readyState === WebSocket.OPEN) {
                    const resizeMessage = `RESIZE:${newCols},${newRows}`;
                    console.log('Sending resize message:', resizeMessage);
                    ws.send(resizeMessage);
                }
                
                // Retry the fitAddon after a short delay
                setTimeout(() => {
                    if (fitAddon) {
                        fitAddon.fit();
                    }
                }, 150);
                
                // Show user feedback
                term.write(`\r\n🔄 Auto-resizing terminal (attempt ${narrowWindowRetryCount}/${MAX_NARROW_WINDOW_RETRIES})...\r\n`);
                
                return; // Don't write the original error message
            } else {
                console.warn('Max auto-resize attempts reached for narrow window issue');
                term.write('\r\n⚠️  Auto-resize failed. Try the "Force Resize" button or manually resize your browser window.\r\n');
            }
        } else {
            // Reset retry count on successful data that doesn't contain errors
            if (narrowWindowRetryCount > 0 && dataString.trim().length > 10 && 
                !dataString.includes('Auto-resizing') && 
                !dataString.includes('🔄') &&
                !containsNarrowError) {
                console.log('Terminal appears to be working normally, resetting retry count');
                narrowWindowRetryCount = 0;
            }
        }
        
        // Write data to terminal
        if (data instanceof Uint8Array) {
            term.write(data);
        } else {
            term.write(data);
        }
    }

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

    const container = document.getElementById('xterm-container');
    container.hidden = false;

    const reconnectButton = document.querySelector('.reconnect-button');
    if (reconnectButton) {
        reconnectButton.hidden = false;
    }

    const resizeButton = document.querySelector('.resize-button');
    if (resizeButton) {
        resizeButton.hidden = false;
    }

    const clearCacheButton = document.querySelector('.clear-cache-button');
    if (clearCacheButton) {
        clearCacheButton.hidden = false;
    }

    const startButton = document.querySelector('.start-button');
    if (startButton) {
        startButton.hidden = true;
    }
}

function forceTerminalResize(extraCols = 0, extraRows = 0) {
    if (!term) {
        console.error('Terminal not initialized');
        return;
    }
    
    const newCols = term.cols + extraCols;
    const newRows = term.rows + extraRows;
    
    console.log(`Manual resize: ${term.cols}x${term.rows} -> ${newCols}x${newRows}`);
    
    term.resize(newCols, newRows);
    
    if (ws && ws.readyState === WebSocket.OPEN) {
        const resizeMessage = `RESIZE:${newCols},${newRows}`;
        ws.send(resizeMessage);
    }
    
    if (fitAddon) {
        setTimeout(() => fitAddon.fit(), 100);
    }
    
    term.write(`\r\n🔧 Manual resize applied: ${newCols}x${newRows}\r\n`);
}

function reconnectTerminal() {
    if (ws) {
        ws.close();
    }
    
    // Reset retry count
    narrowWindowRetryCount = 0;
    
    term.clear();
    term.write('🔄 Reconnecting...\r\n');
    
    // Reinitialize the WebSocket connection
    setTimeout(() => {
        const wsProtocol = process.env.WS_PROTOCOL;
        ws = new WebSocket(`${wsProtocol}://${window.location.host}/ws`);
        
        ws.onopen = function () {
            console.log('WebSocket reconnected');
            term.write('Reconnected to SSH server...\r\n');
            
            if (clientSessionId) {
                const sessionMessage = JSON.stringify({
                    type: 'session',
                    sessionId: clientSessionId
                });
                ws.send(sessionMessage);
                console.log('Sent cached session ID for SSH key reuse');
            }
            
            const cols = term.cols;
            const rows = term.rows;
            ws.send(`RESIZE:${cols},${rows}`);
        };
        
        ws.onmessage = function (event) {
            sshConnected = true;
            let data;
            let dataString = '';

            if (event.data instanceof Blob) {
                const reader = new FileReader();
                reader.onload = function () {
                    const arrayBuffer = reader.result;
                    const uint8Array = new Uint8Array(arrayBuffer);
                    data = uint8Array;
                    dataString = new TextDecoder().decode(uint8Array);
                    handleTerminalData(data, dataString);
                };
                reader.readAsArrayBuffer(event.data);
            } else {
                data = event.data;
                dataString = event.data;
                handleTerminalData(data, dataString);
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
    }, 1000);
}

function clearSSHKeyCache() {
    localStorage.removeItem('sshClientSession');
    clientSessionId = null;
    console.log('SSH key cache cleared');
    updateSSHKeyStatus('SSH key cache cleared - will generate new key on next connection', true);
    
    if (term) {
        term.write('\r\n🗑️  SSH key cache cleared. Reconnect to generate a new key.\r\n');
    }
}

window.initializeTerminal = initializeTerminal;
window.forceTerminalResize = forceTerminalResize;
window.reconnectTerminal = reconnectTerminal;
window.clearSSHKeyCache = clearSSHKeyCache;

window.onload = function () {
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
        };
    }

    // Add event listeners for terminal buttons
    const startButton = document.querySelector('.start-button');
    if (startButton) {
        startButton.addEventListener('click', initializeTerminal);
    }
    
    const reconnectButton = document.querySelector('.reconnect-button');
    if (reconnectButton) {
        reconnectButton.addEventListener('click', reconnectTerminal);
    }
    
    const resizeButton = document.querySelector('.resize-button');
    if (resizeButton) {
        resizeButton.addEventListener('click', () => forceTerminalResize());
    }
    
    const clearCacheButton = document.querySelector('.clear-cache-button');
    if (clearCacheButton) {
        clearCacheButton.addEventListener('click', clearSSHKeyCache);
    }
}