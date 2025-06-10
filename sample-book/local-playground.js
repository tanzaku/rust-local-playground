'use strict';

/* global default_theme, default_dark_theme, default_light_theme, hljs, ClipboardJS */

(function codeSnippets() {
    function fetch_with_timeout(url, options, timeout = 6000) {
        return Promise.race([
            fetch(url, options),
            new Promise((_, reject) => setTimeout(() => reject(new Error('timeout')), timeout)),
        ]);
    }

    const playgrounds = Array.from(document.querySelectorAll('.playground'));

    function run_rust_code(code_block) {
        let result_block = code_block.querySelector('.result');
        if (!result_block) {
            result_block = document.createElement('code');
            result_block.className = 'result hljs language-bash';

            code_block.append(result_block);
        }

        const text = playground_text(code_block);
        const classes = code_block.querySelector('code').classList;
        let edition = '2015';
        classes.forEach(className => {
            if (className.startsWith('edition')) {
                edition = className.slice(7);
            }
        });
        const params = {
            version: 'stable',
            optimize: '0',
            code: text,
            edition: edition,
        };

        if (text.indexOf('#![feature') !== -1) {
            params.version = 'nightly';
        }

        result_block.innerText = 'Running...';

        // fetch_with_timeout('https://play.rust-lang.org/evaluate.json', {
        fetch_with_timeout('http://127.0.0.1:8081/execute', {
            headers: {
                'Content-Type': 'application/json',
            },
            method: 'POST',
            mode: 'cors',
            body: JSON.stringify(params),
        })
            .then(response => response.json())
            .then(response => {
                if (response.result.trim() === '') {
                    result_block.innerText = 'No output';
                    result_block.classList.add('result-no-output');
                } else {
                    result_block.innerText = response.result;
                    result_block.classList.remove('result-no-output');
                }
            })
            .catch(error => result_block.innerText = 'Playground Communication: ' + error.message);
    }

    // Syntax highlighting Configuration
    hljs.configure({
        tabReplace: '    ', // 4 spaces
        languages: [], // Languages used for auto-detection
    });

    const code_nodes = Array
        .from(document.querySelectorAll('code'))
        // Don't highlight `inline code` blocks in headers.
        .filter(function (node) {
            return !node.parentElement.classList.contains('header');
        });

    if (window.ace) {
        // language-rust class needs to be removed for editable
        // blocks or highlightjs will capture events
        code_nodes
            .filter(function (node) {
                return node.classList.contains('editable');
            })
            .forEach(function (block) {
                block.classList.remove('language-rust');
            });

        code_nodes
            .filter(function (node) {
                return !node.classList.contains('editable');
            })
            .forEach(function (block) {
                hljs.highlightBlock(block);
            });
    } else {
        code_nodes.forEach(function (block) {
            hljs.highlightBlock(block);
        });
    }

    // Adding the hljs class gives code blocks the color css
    // even if highlighting doesn't apply
    code_nodes.forEach(function (block) {
        block.classList.add('hljs');
    });

    Array.from(document.querySelectorAll('pre code')).forEach(function (block) {
        const pre_block = block.parentNode;

        let oldButtons = pre_block.querySelector('.buttons');
        if (oldButtons) {
            oldButtons.remove();
        }
    });

    Array.from(document.querySelectorAll('code.hljs')).forEach(function (block) {

        const lines = Array.from(block.querySelectorAll('.boring'));
        // If no lines were hidden, return
        if (!lines.length) {
            return;
        }
        block.classList.add('hide-boring');

        const buttons = document.createElement('div');
        buttons.className = 'buttons';
        buttons.innerHTML = '<button class="fa fa-eye" title="Show hidden lines" \
    aria-label="Show hidden lines"></button>';

        // add expand button
        const pre_block = block.parentNode;

        pre_block.insertBefore(buttons, pre_block.firstChild);

        pre_block.querySelector('.buttons').addEventListener('click', function (e) {
            if (e.target.classList.contains('fa-eye')) {
                e.target.classList.remove('fa-eye');
                e.target.classList.add('fa-eye-slash');
                e.target.title = 'Hide lines';
                e.target.setAttribute('aria-label', e.target.title);

                block.classList.remove('hide-boring');
            } else if (e.target.classList.contains('fa-eye-slash')) {
                e.target.classList.remove('fa-eye-slash');
                e.target.classList.add('fa-eye');
                e.target.title = 'Show hidden lines';
                e.target.setAttribute('aria-label', e.target.title);

                block.classList.add('hide-boring');
            }
        });
    });

    if (window.playground_copyable) {
        Array.from(document.querySelectorAll('pre code')).forEach(function (block) {
            const pre_block = block.parentNode;
            if (!pre_block.classList.contains('playground')) {
                let buttons = pre_block.querySelector('.buttons');
                if (!buttons) {
                    buttons = document.createElement('div');
                    buttons.className = 'buttons';
                    pre_block.insertBefore(buttons, pre_block.firstChild);
                }

                const clipButton = document.createElement('button');
                clipButton.className = 'clip-button';
                clipButton.title = 'Copy to clipboard';
                clipButton.setAttribute('aria-label', clipButton.title);
                clipButton.innerHTML = '<i class="tooltiptext"></i>';

                buttons.insertBefore(clipButton, buttons.firstChild);
            }
        });
    }

    // Process playground code blocks
    Array.from(document.querySelectorAll('.playground')).forEach(function (pre_block) {
        // Add play button
        let buttons = pre_block.querySelector('.buttons');
        if (!buttons) {
            buttons = document.createElement('div');
            buttons.className = 'buttons';
            pre_block.insertBefore(buttons, pre_block.firstChild);
        }

        const runCodeButton = document.createElement('button');
        runCodeButton.className = 'fa fa-play play-button';
        runCodeButton.hidden = true;
        runCodeButton.title = 'Run this code2';
        runCodeButton.setAttribute('aria-label', runCodeButton.title);

        buttons.insertBefore(runCodeButton, buttons.firstChild);
        runCodeButton.addEventListener('click', () => {
            run_rust_code(pre_block);
        });

        if (window.playground_copyable) {
            const copyCodeClipboardButton = document.createElement('button');
            copyCodeClipboardButton.className = 'clip-button';
            copyCodeClipboardButton.innerHTML = '<i class="tooltiptext"></i>';
            copyCodeClipboardButton.title = 'Copy to clipboard';
            copyCodeClipboardButton.setAttribute('aria-label', copyCodeClipboardButton.title);

            buttons.insertBefore(copyCodeClipboardButton, buttons.firstChild);
        }

        const code_block = pre_block.querySelector('code');
        if (window.ace && code_block.classList.contains('editable')) {
            const undoChangesButton = document.createElement('button');
            undoChangesButton.className = 'fa fa-history reset-button';
            undoChangesButton.title = 'Undo changes';
            undoChangesButton.setAttribute('aria-label', undoChangesButton.title);

            buttons.insertBefore(undoChangesButton, buttons.firstChild);

            undoChangesButton.addEventListener('click', function () {
                const editor = window.ace.edit(code_block);
                editor.setValue(editor.originalCode);
                editor.clearSelection();
            });
        }
    });
})();
