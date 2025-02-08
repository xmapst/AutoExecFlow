let highestZIndex = 1000;
require.config({ paths: { 'vs': baseUrl+basePath+'/vs' } });
const taskTpl = `# 描述
desc: 这是一段任务描述
# 允许节点, 可选, 默认为当前节点
#node: node-01
# 异步执行, 可选, 默认并行,自定义编排时需要设置为true
async: true
# 禁用, 可选, 默认false
#disable: false
# 超时时间, 可选, 默认48小时
timeout: 2m
# 全局环境变量, 可选
env:
  - name: Test
    value: "test_env"
# 步骤列表, 不能为空
step:
    # 步骤名称, 唯一, 可选[当自定义编排是必须设置], 默认自动生成
  - name: 步骤2
    # 描述
    desc: 这是一段步骤描述
    # 超时时间, 可选, 默认任务级超时时间
    timeout: 2m
    # 禁用, 可选, 默认false
    #disable: false
    # 依赖步骤, 可选[自定义编排时用到]
    depends:
      - 步骤1
    # 局部环境变量, 会覆盖同名的全局变量
    env:
      - name: Test
        value: "test_env"
    # 类型
    type: sh
    # 内容
    content: |-
      ping 1.1.1.1
  - name: 步骤1
    # 描述
    desc: 这是一段步骤描述
    timeout: 2m
    env:
      - name: Test
        value: "test_env"
    type: sh
    content: |-
      ping 1.1.1.1
`

G6.registerNode(
    'custom-node',
    {
        drawShape: function drawShape(cfg, group) {
            const color = Utils.getStatusColor(cfg.step.state || 'unknown');
            const r = 3;

            group.addShape('rect', {
                attrs: {
                    x: 0,
                    y: 0,
                    width: 220,
                    height: 70,
                    stroke: color,
                    fill: "#FFFFFF",
                    radius: r,
                },
                name: 'main-box',
                draggable: true,
            });
            group.addShape('rect', {
                attrs: {
                    x: 0,
                    y: 0,
                    width: 220,
                    height: 20,
                    fill: color,
                    radius: [r, r, 0, 0],
                },
                name: 'title-box',
                draggable: true,
            });
            group.addShape('image', {
                attrs: {
                    x: 4,
                    y: 2,
                    height: 16,
                    width: 16,
                    cursor: 'pointer',
                    img: `${baseUrl+basePath}/img/node-icon.png`,
                },
                name: 'node-icon',
            });
            group.addShape('text', {
                attrs: {
                    textBaseline: 'top',
                    y: 6,
                    x: 24,
                    lineHeight: 16,
                    text: cfg.step.name,
                    fill: '#fff',
                },
                name: 'title-text',
            });
            if (cfg.step.code !== undefined) {
                group.addShape('text', {
                    attrs: {
                        textBaseline: 'top',
                        y: 25,
                        x: 6,
                        lineHeight: 3,
                        text: "状态码: " + cfg.step.code,
                        fill: 'rgba(0,0,0, 0.4)',
                    },
                    name: 'title-code',
                });
            }
            group.addShape('text', {
                attrs: {
                    textBaseline: 'top',
                    y: 25,
                    x: 100,
                    lineHeight: 3,
                    text: "状态: " + cfg.step.state,
                    fill: 'rgba(0,0,0, 0.4)',
                },
                name: 'title-state',
            });
            group.addShape('text', {
                attrs: {
                    textBaseline: 'top',
                    y: 40,
                    x: 6,
                    lineHeight: 3,
                    text: cfg.step.time.start ? "开始时间: " + cfg.step.time.start : '开始时间: ---',
                    fill: 'rgba(0,0,0, 0.4)',
                },
                name: 'title-time-start',
            });
            group.addShape('text', {
                attrs: {
                    textBaseline: 'top',
                    y: 55,
                    x: 6,
                    lineHeight: 3,
                    text: cfg.step.time.end ? "结束时间: " + cfg.step.time.end : '结束时间: ---',
                    fill: 'rgba(0,0,0, 0.4)',
                },
                name: 'title-time-end',
            });
            return group;
        },
    },
);

// 工具类
class Utils {
    static removeElementById(id) {
        const element = document.getElementById(id);
        if (element) element.remove();
    }

    static getStatusColor(status) {
        const statusColorMap = {
            'running': '#00BFFF',
            'stopped': '#32CD32',
            'failed': '#FF6B6B',
            'pending': '#FFA500',
            'timeout': '#FFC107',
            'canceled': '#9B59B6',
            'skipped': '#007BFF',
            'blacked': '#000000',
            'unknown': '#A9A9A9'
        };
        return statusColorMap[status] || 'gray';
    }

    static escapeHTML(html) {
        const div = document.createElement('div');
        div.textContent = html;
        return div.innerHTML;
    }

    static taskManager(taskName, action) {
        const confirmed = confirm(`确定要${action} "${taskName}"?`);
        if (confirmed) {
            fetch(`${baseUrl}${taskUrl}/${taskName}?action=${action}`, {
                method: 'PUT',
            })
                .then(response => {
                    if (!response.ok) {
                        throw new Error('Network response was not ok');
                    }
                    return response.json();
                }).catch(error => {
                console.log('There was a problem with the fetch operation:', error);
                throw error;
            });
        }
    }

    static stepManager(taskName, stepName, action) {
        const confirmed = confirm(`确定要${action} "${stepName}"?`);
        if (confirmed) {
            fetch(`${baseUrl}${taskUrl}/${taskName}/step/${stepName}?action=${action}`, {
                method: 'PUT',
            })
                .then(response => {
                    if (!response.ok) {
                        throw new Error('Network response was not ok');
                    }
                    return response.json();
                }).catch(error => {
                console.log('There was a problem with the fetch operation:', error);
                throw error;
            });
        }
    }

    static InitializeEditor(eElementID, data, readOnly = false) {
        // 返回一个 Promise
        return new Promise((resolve) => {
            require(['vs/editor/editor.main'], () => {
                const editor = monaco.editor.create(document.getElementById(eElementID), {
                    value: data,
                    language: 'yaml',
                    theme: 'vs-dark',
                    readOnly: readOnly,
                    autoIndent: true,
                    automaticLayout: true,
                    overviewRulerBorder: false,
                    foldingStrategy: 'indentation',
                    lineNumbers: 'on',
                    minimap: { enabled: false },
                    tabSize: 4,
                    mouseWheelZoom: true,
                    formatOnType: true,
                    formatOnPaste: true,
                    cursorStyle: 'line',
                    fontSize: 12,
                });
                resolve(editor); // 初始化后返回 editor 实例
            });
        });
    }

}

// 通用的WebSocket管理器
class WebSocketManager {
    constructor(url, onMessage, onError, reconnectInterval = 5000) {
        this.url = url;
        this.onMessage = onMessage;
        this.onError = onError;
        this.reconnectInterval = reconnectInterval;
        this.socket = null;
        this.isManuallyClosed = false;
        this.connect();
    }

    connect() {
        if (this.socket && (this.socket.readyState === WebSocket.OPEN || this.socket.readyState === WebSocket.CONNECTING)) {
            console.log("WebSocket is already open or connecting.");
            return;
        }
        this.isManuallyClosed = false;
        this.socket = new WebSocket(this.url);

        this.socket.onopen = () => {
            console.log("WebSocket connection established.");
        };

        this.socket.onmessage = (event) => {
            const data = JSON.parse(event.data);
            this.onMessage(data);
        };

        this.socket.onerror = (error) => {
            console.error('WebSocket error:', error);
            if (this.onError) this.onError(error);
            this.reconnect();
        };

        this.socket.onclose = (event) => {
            if (event.wasClean) {
                console.log("WebSocket closed cleanly, no reconnection.");
            } else {
                console.log(`WebSocket closed with code: ${event.code}, reason: ${event.reason}`);
                this.reconnect();
            }
        };
    }

    reconnect() {
        if (this.isManuallyClosed) return;
        setTimeout(() => {
            console.log("Attempting to reconnect...");
            this.connect();
        }, this.reconnectInterval);
    }

    send(data) {
        if (this.socket && this.socket.readyState === WebSocket.OPEN) {
            this.socket.send(JSON.stringify(data));
        } else {
            console.warn("WebSocket is not open. Unable to send data.");
        }
    }

    close() {
        this.isManuallyClosed = true;
        if (this.socket) {
            this.socket.close(1000, "Normal closure");
        }
    }
}

class TaskModal {
    constructor(taskName) {
        this.webSocketManager = null;
        this.taskName = taskName;
        this.task = null;
        this.graph = null;
        this.stepModals = [];
        this.init(taskName);
    };

    init(taskName) {
        // 获取任务详细
        fetch(`${baseUrl}${taskUrl}/${taskName}`)
            .then(response => response.json())
            .then(res => {
                this.task = res.data;
               this.start(taskName);
            }).catch(error => {
                console.log('There was a problem with the fetch operation:', error);
                throw error;
        })
    };

    start(taskName) {
        this.webSocketManager = new WebSocketManager(`${wsBaseUrl}${taskUrl}/${taskName}/step`, this.updateGraphData, (error) => {
            alert('WebSocket error: ' + error);
            this.closeModal();
        })
        this.createModal();
        this.addEventListeners();
    }

    createModal() {
        Utils.removeElementById('task-card');
        Utils.removeElementById("task-modal-overlay");

        const overlay = document.createElement('div');
        overlay.setAttribute("id", "task-modal-overlay");
        overlay.className = 'modal-overlay';
        document.body.appendChild(overlay);

        const card = document.createElement('div');
        card.setAttribute("id", "task-card");
        card.className = 'card-one';
        card.style.left = "30%";
        card.innerHTML = `
                <div class="card-header">
                    <h5 style="margin-left: 15px;">名称: ${this.task.name}</h5>
                    <span class="card-close" id="close-task-card">&times;</span >
                </div>
                <hr>
                <div id="task-card-body" class="card-body">
                    ${this.task.env ? `
                     <div id="task-card-right" class="card-body-left">
                         <h5>环境变量:</h5>
                         <div id="${this.task.name + '-env'}">
                             <pre class="env">${this.task.env.map(env => `- name: ${env.name}\n  value: ${env.value}`).join('\n')}</pre>
                         </div>
                     </div>
                    ` : ''}
                    <div id="task-card-left" class="card-body-right"></div>
                </div>
            `;

        document.body.appendChild(card);

        setTimeout(() => {
            overlay.classList.add('show');
            card.classList.add('show');
        }, 10);
    };

    createGraph(data) {
        const card = document.getElementById('task-card-left');
        const width = card.scrollWidth ;
        const height = card.scrollHeight || 500;
        const grid = new G6.Grid();
        const menu = new G6.Menu({
            itemTypes: ['node'],
            getContent(e) {
                const code = e.item.getModel().code;
                const step = e.item.getModel().step;
                // 判断节点是否找到
                if (!step || code === 0 || code === 1002 || code === 1003 ) {
                    return '无操作可选';
                }

                // 根据状态生成菜单项
                let menuContent = '';
                if (step.state === 'running') {
                    menuContent += '<a href="#" id="kill-step">强杀</a>';
                } else if (step.state === 'paused') {
                    menuContent += '<a href="#" id="resume-step">解挂</a>';
                    menuContent += '<a href="#" id="kill-step">强杀</a>';
                } else if (step.state === 'pending') {
                    menuContent += '<a href="#" id="pause-step">挂起</a>';
                    menuContent += '<a href="#" id="kill-step">强杀</a>';
                }
                if (menuContent === '') {
                    return '无操作可选';
                }
                return menuContent
            },
            handleMenuClick(target, item) {
                const taskName = item.getModel().taskName;
                const step = item.getModel().step;
                // 根据点击的菜单项执行相应的操作
                if (target.id === 'kill-step') {
                    Utils.stepManager(taskName, step.name, 'kill')
                } else if (target.id === 'pause-step') {
                    Utils.stepManager(taskName, step.name, 'pause')
                } else if (target.id === 'resume-step') {
                    Utils.stepManager(taskName, step.name, 'resume')
                }
            },
        });
        const toolbar = new G6.ToolBar({
            getContent: () => {
                return `<ul class="g6-component-toolbar" style="top: 0; left: 1011px;">
                    <li code="zoomOut">
                        <svg class="icon" viewBox="0 0 1024 1024" version="1.1" xmlns="http://www.w3.org/2000/svg" width="24" height="24">
                            <path d="M658.432 428.736a33.216 33.216 0 0 1-33.152 33.152H525.824v99.456a33.216 33.216 0 0 1-66.304 0V461.888H360.064a33.152 33.152 0 0 1 0-66.304H459.52V296.128a33.152 33.152 0 0 1 66.304 0V395.52H625.28c18.24 0 33.152 14.848 33.152 33.152z m299.776 521.792a43.328 43.328 0 0 1-60.864-6.912l-189.248-220.992a362.368 362.368 0 0 1-215.36 70.848 364.8 364.8 0 1 1 364.8-364.736 363.072 363.072 0 0 1-86.912 235.968l192.384 224.64a43.392 43.392 0 0 1-4.8 61.184z m-465.536-223.36a298.816 298.816 0 0 0 298.432-298.432 298.816 298.816 0 0 0-298.432-298.432A298.816 298.816 0 0 0 194.24 428.8a298.816 298.816 0 0 0 298.432 298.432z"></path>
                        </svg>
                    </li>
                    <li code="zoomIn">
                        <svg class="icon" viewBox="0 0 1024 1024" version="1.1" xmlns="http://www.w3.org/2000/svg" width="24" height="24">
                            <path d="M639.936 416a32 32 0 0 1-32 32h-256a32 32 0 0 1 0-64h256a32 32 0 0 1 32 32z m289.28 503.552a41.792 41.792 0 0 1-58.752-6.656l-182.656-213.248A349.76 349.76 0 0 1 480 768 352 352 0 1 1 832 416a350.4 350.4 0 0 1-83.84 227.712l185.664 216.768a41.856 41.856 0 0 1-4.608 59.072zM479.936 704c158.784 0 288-129.216 288-288S638.72 128 479.936 128a288.32 288.32 0 0 0-288 288c0 158.784 129.216 288 288 288z" p-id="3853"></path>
                        </svg>
                    </li>
                    <li code="realZoom">
                        <svg class="icon" viewBox="0 0 1024 1024" version="1.1" xmlns="http://www.w3.org/2000/svg" width="20" height="24">
                            <path d="M384 320v384H320V320h64z m256 0v384H576V320h64zM512 576v64H448V576h64z m0-192v64H448V384h64z m355.968 576H92.032A28.16 28.16 0 0 1 64 931.968V28.032C64 12.608 76.608 0 95.168 0h610.368L896 192v739.968a28.16 28.16 0 0 1-28.032 28.032zM704 64v128h128l-128-128z m128 192h-190.464V64H128v832h704V256z"></path>
                        </svg>
                    </li>
                    <li code="autoZoom">
                        <svg class="icon" viewBox="0 0 1024 1024" version="1.1" xmlns="http://www.w3.org/2000/svg" width="20" height="24">
                            <path d="M684.288 305.28l0.128-0.64-0.128-0.64V99.712c0-19.84 15.552-35.904 34.496-35.712a35.072 35.072 0 0 1 34.56 35.776v171.008h170.944c19.648 0 35.84 15.488 35.712 34.432a35.072 35.072 0 0 1-35.84 34.496h-204.16l-0.64-0.128a32.768 32.768 0 0 1-20.864-7.552c-1.344-1.024-2.816-1.664-3.968-2.816-0.384-0.32-0.512-0.768-0.832-1.088a33.472 33.472 0 0 1-9.408-22.848zM305.28 64a35.072 35.072 0 0 0-34.56 35.776v171.008H99.776A35.072 35.072 0 0 0 64 305.216c0 18.944 15.872 34.496 35.84 34.496h204.16l0.64-0.128a32.896 32.896 0 0 0 20.864-7.552c1.344-1.024 2.816-1.664 3.904-2.816 0.384-0.32 0.512-0.768 0.768-1.088a33.024 33.024 0 0 0 9.536-22.848l-0.128-0.64 0.128-0.704V99.712A35.008 35.008 0 0 0 305.216 64z m618.944 620.288h-204.16l-0.64 0.128-0.512-0.128c-7.808 0-14.72 3.2-20.48 7.68-1.28 1.024-2.752 1.664-3.84 2.752-0.384 0.32-0.512 0.768-0.832 1.088a33.664 33.664 0 0 0-9.408 22.912l0.128 0.64-0.128 0.704v204.288c0 19.712 15.552 35.904 34.496 35.712a35.072 35.072 0 0 0 34.56-35.776V753.28h170.944c19.648 0 35.84-15.488 35.712-34.432a35.072 35.072 0 0 0-35.84-34.496z m-593.92 11.52c-0.256-0.32-0.384-0.768-0.768-1.088-1.088-1.088-2.56-1.728-3.84-2.688a33.088 33.088 0 0 0-20.48-7.68l-0.512 0.064-0.64-0.128H99.84a35.072 35.072 0 0 0-35.84 34.496 35.072 35.072 0 0 0 35.712 34.432H270.72v171.008c0 19.84 15.552 35.84 34.56 35.776a35.008 35.008 0 0 0 34.432-35.712V720l-0.128-0.64 0.128-0.704a33.344 33.344 0 0 0-9.472-22.848zM512 374.144a137.92 137.92 0 1 0 0.128 275.84A137.92 137.92 0 0 0 512 374.08z"></path>
                        </svg>
                    </li>
                </ul>`
            },
            position: {
                x: width/2,
                y: 0,
            },
        });
        this.graph  = new G6.Graph({
            container: "task-card-left",
            height: height,
            width: width,
            layout: {
                type: 'dagre',
                rankdir: 'LR',
                align: 'UL',
                nodesep: 24,
                ranksep: 78,
            },
            modes: {
                default: ['drag-canvas', 'drag-node', 'zoom-canvas','activate-relations'],
            },
            defaultNode: {
                type: 'custom-node',
                anchorPoints: [
                    [0, 0.5],
                    [0.5, 0],
                    [0.5, 1],
                    [1, 0.5],
                ],
            },
            defaultEdge: {
                type: 'polyline',
                style: {
                    endArrow: {
                        path: G6.Arrow.triangle(4, 4, 10),
                        d: 10
                    },
                    stroke: '#F6BD16',
                },
            },
            plugins: [grid, menu, toolbar],
        });

        this.graph.on('node:click', evt => {
            const step = evt.item.getModel().step;
            this.openStepModal(step.name);
        });
        this.graph.on('canvas:click', () => {
            this.closeAllStepModals();
        });
        this.graph.data(data);
        this.graph.render();
    };

    updateGraphData = (res) => {
        if (!res.data) {
            return;
        }
        if (this.graph) {
            res.data.forEach(step => {
                let node = {
                    id: step.name,
                    step: step,
                    code: res.code,
                    taskName: this.taskName,
                };
                const item = this.graph.findById(node.id);
                if (item) {
                    item.step = node.step;
                    item.taskName = this.taskName;
                    item.code = res.code;
                    this.graph.updateItem(node.id, item);
                } else {
                    this.graph.addItem('node', node);
                }

                if (step.depends) {
                    step.depends.forEach(depend => {
                        let edge = {
                            id: depend + '-' + step.name,
                            source: depend,
                            target: step.name,
                        };
                        const edgeItem = this.graph.findById(edge.id);
                        if (!edgeItem) {
                            this.graph.addItem('edge', edge);
                        }
                    })
                }
            });
            return;
        }
        let data = {
            nodes: [],
            edges: [],
        };
        res.data.forEach(step => {
            data.nodes.push({
                id: step.name,
                step: step,
                code: res.code,
                taskName: this.taskName,
            });
            if (step.depends) {
                step.depends.forEach(depend => {
                    let edge = {
                        id: depend + '-' + step.name,
                        source: depend,
                        target: step.name,
                    };
                    data.edges.push(edge);
                });
            }
        });
        this.createGraph(data);
    };

    closeAllStepModals() {
        highestZIndex = 1000;
        this.stepModals.forEach(stepModal => stepModal.closeModal());
        this.stepModals = [];
    }

    openStepModal(stepName) {
        const stepModal = new StepModal(this.task.name, stepName);
        this.stepModals.push(stepModal);
    }

    addEventListeners() {
        document.getElementById('close-task-card').addEventListener('click', () => this.closeModal());
        document.getElementById("task-modal-overlay").addEventListener("click", () => this.closeModal());
        window.addEventListener('keydown', this.handleEscapeKey.bind(this));
    };

    handleEscapeKey(event) {
        if (event.key === 'Escape') {
            this.closeModal();
        }
    }

    closeModal() {
        if (this.webSocketManager) {
            this.webSocketManager.close();
            this.webSocketManager = null;
        }
        this.closeAllStepModals();
        const card = document.getElementById("task-card");
        const overlay = document.getElementById("task-modal-overlay");
        card.classList.remove('show');
        overlay.classList.remove('show');
        setTimeout(() => {
            Utils.removeElementById("task-card");
            Utils.removeElementById("task-modal-overlay");
        }, 300);
        window.removeEventListener('keydown', this.handleEscapeKey.bind(this));
    };
}

class StepModal {
    constructor(taskName, stepName) {
        this.webSocketManager = null;
        this.step = null;
        this.isDragging = false;
        this.offsetX = 0;
        this.offsetY = 0;
        this.init(taskName, stepName);
    };

    init(taskName,stepName) {
        // 获取任务详细
        fetch(`${baseUrl}${taskUrl}/${taskName}/step/${stepName}`)
            .then(response => response.json())
            .then(res => {
                this.step = res.data;
                this.start(taskName, stepName);
            }).catch(error => {
            console.log('There was a problem with the fetch operation:', error);
            throw error;
        })
    };

    start(taskName, stepName) {
        const existingCard = document.getElementById(this.step.name + "-step-card");
        if (existingCard) {
            existingCard.style.zIndex = ++highestZIndex;
            return;
        }
        this.webSocketManager = new WebSocketManager(`${wsBaseUrl}${taskUrl}/${taskName}/step/${stepName}/log`,this.updateStepOutput, ()=> {
            const outputElement = document.getElementById('step-output-text');
            outputElement.innerHTML = `<pre class="step-card-code">${Utils.escapeHTML(this.step.message)}</pre>`;
        });
        this.createModal();
        this.addEventListeners();
    };

    createModal() {
        const card = document.createElement('div');
        card.setAttribute("id", this.step.name + "-step-card");
        card.className = 'step-card';
        card.style.zIndex = ++highestZIndex;
        card.innerHTML = `
                <div id="${this.step.name + '-step-card-header'}" class="card-header">
                    <h5>名称: ${this.step.name}</h3>
                    <h5>类型: ${this.step.type}</h5>
                    <span class="card-close" id="${this.step.name + '-close-step-card'}">&times;</span>
                </div>
                <hr>
                <div id="${this.step.name + '-step-card-body'}" class="card-body">
                    ${this.step.env ? `
                        <div id="${this.step.name + '-step-card-left'}" class="card-body-left">
                            <h5>环境变量:</h5>
                            <div id="${this.step.name + '-env'}">
                                <pre class="env">${this.step.env.map(env => `- name: ${env.name}\n  value: ${env.value}`).join('\n')}</pre>
                            </div>
                        </div>
                    ` : ''}
                    <div id="${this.step.name + '-step-card-right'}" class="card-body-right">
                        <h5>输入: </h5>
                        <div id="${this.step.name + '-step-input'}" class="step-card-input">
                            <pre id="${this.step.name + '-step-input-text'}" class="step-card-code"></pre>
                        </div>
                        <h5>输出: </h5>
                        <div id="${this.step.name + '-step-output'}" class="step-card-output">
                            <pre id="${this.step.name + '-step-output-text'}" class="step-card-code"></pre>
                        </div>
                    </div>
                </div>
            `;
        document.body.appendChild(card);

        const inputDocument = document.getElementById(this.step.name + '-step-input-text');
        inputDocument.innerText = this.step.content;

        const header = document.getElementById(this.step.name + '-step-card-header');
        this.addDragEventListeners(header, card);
        card.addEventListener('mousedown', () => {
            this.bringToFront(card);
        });
    };

    updateStepOutput = (res) => {
        const outputElement = document.getElementById(this.step.name + '-step-output-text');

        if (outputElement) {
            if (!res.data) {
                return;
            }
            let scrollTop = outputElement.scrollTop;
            let data = [];
            res.data.forEach(item => {
                data.push(item.content);
            });
            outputElement.textContent += data.join("\n");
            outputElement.textContent += "\n";
            outputElement.scrollTop = scrollTop;
        }
    };

    addEventListeners() {
        document.getElementById(this.step.name + '-close-step-card').addEventListener('click', () => this.closeModal());
    };

    closeModal() {
        if (this.webSocketManager) {
            this.webSocketManager.close();
            this.webSocketManager = null;
        }
        Utils.removeElementById(this.step.name + "-step-card");
    };

    addDragEventListeners(header, card) {
        header.style.cursor = 'move';

        header.addEventListener('mousedown', (e) => {
            this.isDragging = true;
            this.offsetX = e.clientX - card.offsetLeft;
            this.offsetY = e.clientY - card.offsetTop;
            document.addEventListener('mousemove', this.handleMouseMove);
            document.addEventListener('mouseup', this.handleMouseUp);
        });
    }

    handleMouseMove = (e) => {
        if (this.isDragging) {
            const card = document.getElementById(this.step.name + "-step-card");
            card.style.left = e.clientX - this.offsetX + 'px';
            card.style.top = e.clientY - this.offsetY + 'px';
        }
    }

    handleMouseUp = () => {
        this.isDragging = false;
        document.removeEventListener('mousemove', this.handleMouseMove);
        document.removeEventListener('mouseup', this.handleMouseUp);
    }

    bringToFront(card) {
        card.style.zIndex = ++highestZIndex;
    }
}

class TaskAddCard {
    constructor() {
        this.editor = null
        this.show();
    }

    show() {
        Utils.removeElementById("add-task-card");
        Utils.removeElementById("task-add-modal-overlay");

        const overlay = document.createElement('div');
        overlay.setAttribute("id", "task-add-modal-overlay");
        overlay.className = 'modal-overlay';
        document.body.appendChild(overlay);

        const card = document.createElement('div');
        card.setAttribute("id", "add-task-card");
        card.className = 'card-one';
        card.innerHTML = `
            <div class="card-header">
                <div class="button" style="position: fixed; top: 6px; right: 12px;">
                    <button id="create-task" class="button-sure">创建</button>
                    <button id="cancel-task" class="button-cancel">取消</button>
                </div>
            </div>
            <div class="create-content">
                <div id="yaml-editor"></div>
            </div>
        `;
        document.body.appendChild(card);

        setTimeout(() => {
            overlay.classList.add('show');
            card.classList.add('show');
        }, 10);

        Utils.InitializeEditor('yaml-editor', "# 任务名称, 可选, 默认自动生成\nname: 测试\n"+taskTpl).then(editor => this.editor = editor);
        this.bindEvents();
    }

    closeModal() {
        const card = document.getElementById("add-task-card");
        const overlay = document.getElementById("task-add-modal-overlay");
        card.classList.remove('show');
        overlay.classList.remove('show');

        setTimeout(() => {
            Utils.removeElementById("add-task-card");
            Utils.removeElementById("task-add-modal-overlay");
        }, 300);
    }

    bindEvents() {
        document.getElementById("create-task").addEventListener("click", () => this.createTask());
        document.getElementById("cancel-task").addEventListener("click", () => this.closeModal());
        document.getElementById("task-add-modal-overlay").addEventListener("click", () => this.closeModal());
    }

    createTask() {
        const yamlContent = this.editor.getValue().trim();
        if (yamlContent === "") {
            alert("请输入YAML内容");
            return;
        }
        try {
            fetch(`${baseUrl}${taskUrl}`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/yaml',
                },
                body: yamlContent,
            })
                .then(response => response.json())
                .then(res => {
                    if (res.code === 0) {
                        new TaskModal(res.data.name)
                    } else {
                        alert("任务添加失败: " + Utils.escapeHTML(res.message));
                    }
                });
        } catch (e) {
            alert("Error: " + e.message);
        }
        this.closeModal();
    }
}

class PipelineAddCard {
    constructor() {
        this.editor = null;
        this.show();
    }

    show() {
        Utils.removeElementById("add-pipeline-card");
        Utils.removeElementById("add-pipeline-modal-overlay");

        const overlay = document.createElement('div');
        overlay.setAttribute("id", "add-pipeline-modal-overlay");
        overlay.className = 'modal-overlay';
        document.body.appendChild(overlay);

        const card = document.createElement('div');
        card.setAttribute("id", "add-pipeline-card");
        card.className = 'card-one';
        card.innerHTML = `
            <div class="card-header">
                <div class="button" style="position: fixed; top: 6px; right: 12px;">
                    <button id="create-pipeline" class="button-sure">创建</button>
                    <button id="cancel-create-pipeline" class="button-cancel">取消</button>
                </div>
            </div>
            <div class="create-content">
                <div style="width: 100%">
                    <label for="name">名称:</label>
                    <input type="text" id="name" name="name" style="width: 200px">
                    <div style="position: absolute; display: contents">
                        <label for="tplType">模板类型:</label>
                        <select id="tplType" name="tplType">
                            <option value="jinja2">jinja2</option>
                            <!-- 如果有其他选项，也可以在这里添加 -->
                        </select>
                    </div>
                </div>
                
                <div>
                    <label for="description">描述:</label>
                    <textarea id="description" name="description" style="width: 100%; height: 100px; resize: none;"></textarea>
                </div>
                <div style="position: absolute; margin-bottom: 0;height: 100%; width: 100%">
                    <label for="name">内容:</label>
                    <div id="yaml-editor"></div>
                </div>
            </div>
        `;
        document.body.appendChild(card);

        setTimeout(() => {
            overlay.classList.add('show');
            card.classList.add('show');
        }, 10);

        Utils.InitializeEditor('yaml-editor', taskTpl).then(editor => this.editor = editor);
        this.bindEvents();
    }

    closeModal() {
        const card = document.getElementById("add-pipeline-card");
        const overlay = document.getElementById("add-pipeline-modal-overlay");
        card.classList.remove('show');
        overlay.classList.remove('show');

        setTimeout(() => {
            Utils.removeElementById("add-pipeline-card");
            Utils.removeElementById("add-pipeline-modal-overlay");
        }, 300);
    }

    bindEvents() {
        document.getElementById("create-pipeline").addEventListener("click", () => this.createPipeline());
        document.getElementById("cancel-create-pipeline").addEventListener("click", () => this.closeModal());
        document.getElementById("add-pipeline-modal-overlay").addEventListener("click", () => this.closeModal());
    }

    createPipeline() {
        const name = document.getElementById("name").value.trim();
        let description = "";
        const descriptionElement = document.getElementById("description");
        if (descriptionElement) {
            description = descriptionElement.value.trim();
        }
        const tplType = document.getElementById("tplType").value;
        const content = this.editor.getValue().trim();
        if (name === "") {
            alert("名称不能为空！");
            return; // 阻止提交
        }
        if (content === "") {
            alert("请输入YAML内容");
            return;
        }

        const escapedContent = content
            .replace(/\n/g, '\n    '); // 在每行前加空格，保持缩进
        const escapedDescription = description
            .replace(/\n/g, '\n    '); // 在每行前加空格，保持缩进
        try {
            fetch(`${baseUrl}${pipelineUrl}`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/yaml',
                },
                body: `name: ${name}\ndesc: |-\n    ${escapedDescription}\ntplType: ${tplType}\ncontent: |-\n    ${escapedContent}
                `,
            })
                .then(response => {
                    if (!response.ok) {
                        throw new Error(`HTTP error! Status: ${response.status}`);
                    }
                    return response.json();
                })
                .then(data => {
                    if (data.code === 0) {
                        alert("流水线添加成功");
                    } else {
                        alert("流水线添加失败: " + Utils.escapeHTML(data.message));
                    }
                });
        } catch (e) {
            alert("Error: " + e.message);
            this.closeModal();
        }
    }
}

class PipelineEditCard {
    constructor(pipelineName) {
        this.pipelineName = pipelineName;
        this.pipeline = null;
        this.editor = null;
        this.getPipelineData().then(r => {
            this.pipeline = r;
            this.init();
        });
    }

    getPipelineData() {
        return new Promise((resolve, reject) => {
            fetch(`${baseUrl}${pipelineUrl}/${this.pipelineName}`)
                .then(response => response.json())
                .then(res => {
                    if (res.code === 0) {
                        resolve(res.data);
                    } else {
                        reject(new Error(res.message));
                    }
                })
                .catch(error => reject(error));
        });
    }

    init() {
        Utils.removeElementById("edit-pipeline-card");
        Utils.removeElementById("edit-pipeline-modal-overlay");

        const overlay = document.createElement('div');
        overlay.setAttribute("id", "edit-pipeline-modal-overlay");
        overlay.className = 'modal-overlay';
        document.body.appendChild(overlay);

        const card = document.createElement('div');
        card.setAttribute("id", "edit-pipeline-card");
        card.className = 'card-one';
        card.innerHTML = `
            <div class="card-header">
                <div class="button" style="position: fixed; top: 6px; right: 12px;">
                    <button id="edit-post-pipeline" class="button-sure">提交</button>
                    <button id="cancel-edit-pipeline" class="button-cancel">取消</button>
                </div>
            </div>
            <div class="create-content">
                <div>
                    <label for="name">名称:</label>
                    <input type="text" id="name" name="name" style="width: 100%" value="${this.pipeline.name}" readonly>
                </div>
                
                <div>
                    <label for="description">描述:</label>
                    <textarea id="description" name="description" style="width: 100%; height: 100px; resize: none;">${this.pipeline.desc ? this.pipeline.desc : ''}</textarea>
                </div>
                <div style="position: absolute; margin-bottom: 0;height: 100%; width: 100%">
                    <label for="name">内容:</label>
                    <div id="yaml-editor"></div>
                </div>
            </div>
        `;
        document.body.appendChild(card);

        setTimeout(() => {
            overlay.classList.add('show');
            card.classList.add('show');
        }, 10);

        Utils.InitializeEditor('yaml-editor', this.pipeline.content).then(editor => this.editor = editor);
        this.bindEvents();
    }

    closeModal() {
        const card = document.getElementById("edit-pipeline-card");
        const overlay = document.getElementById("edit-pipeline-modal-overlay");
        card.classList.remove('show');
        overlay.classList.remove('show');

        setTimeout(() => {
            Utils.removeElementById("edit-pipeline-card");
            Utils.removeElementById("edit-pipeline-modal-overlay");
        }, 300);
    }

    bindEvents() {
        document.getElementById("edit-post-pipeline").addEventListener("click", () => this.editPipeline());
        document.getElementById("cancel-edit-pipeline").addEventListener("click", () => this.closeModal());
        document.getElementById("edit-pipeline-modal-overlay").addEventListener("click", () => this.closeModal());
    }

    editPipeline() {
        let description = "";
        const descriptionElement = document.getElementById("description");
        if (descriptionElement) {
            description = descriptionElement.value.trim();
        }
        const content = this.editor.getValue().trim();
        if (content === "") {
            alert("请输入YAML内容");
            return;
        }

        const escapedContent = content
            .replace(/\n/g, '\n    '); // 在每行前加空格，保持缩进
        const escapedDescription = description
            .replace(/\n/g, '\n    '); // 在每行前加空格，保持缩进
        try {
            fetch(`${baseUrl}${pipelineUrl}/${this.pipelineName}`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/yaml',
                },
                body: `desc: |-\n    ${escapedDescription}\ntplType: ${this.pipeline.tplType}\ncontent: |-\n    ${escapedContent}
                `,
            })
                .then(response => {
                    if (!response.ok) {
                        throw new Error(`HTTP error! Status: ${response.status}`);
                    }
                    return response.json();
                })
                .then(res => {
                    if (res.code === 0) {
                        alert("流水线编辑成功");
                    } else {
                        alert("流水线编辑失败: " + Utils.escapeHTML(res.message));
                    }
                    this.closeModal();
                });
        } catch (e) {
            alert("Error: " + e.message);
            this.closeModal();
        }
    }
}

let paramTpl = `params: \n  imageTag: 111`

class RunPipelineModal {
    constructor(pipelineName) {
        this.editor = null;
        this.pipelineName = pipelineName;
        this.init();
    }

    init() {
        Utils.removeElementById("run-pipeline-card");
        Utils.removeElementById("run-pipeline-modal-overlay");

        const overlay = document.createElement('div');
        overlay.setAttribute("id", "run-pipeline-modal-overlay");
        overlay.className = 'modal-overlay';
        document.body.appendChild(overlay);

        const card = document.createElement('div');
        card.setAttribute("id", "run-pipeline-card");
        card.className = 'card-one';
        card.innerHTML = `
            <div class="card-header">
                <div class="button" style="position: fixed; top: 6px; right: 12px;">
                    <button id="run-start-pipeline" class="button-sure">执行</button>
                    <button id="cancel-run-pipeline" class="button-cancel">取消</button>
                </div>
            </div>
            <div class="create-content">
                <label for="name">参数:</label>
                <div id="yaml-editor"></div>
            </div>
        `;
        document.body.appendChild(card);

        setTimeout(() => {
            overlay.classList.add('show');
            card.classList.add('show');
        }, 10);

        Utils.InitializeEditor('yaml-editor', paramTpl).then(editor => this.editor = editor);
        this.bindEvents();
    }

    closeModal() {
        const card = document.getElementById("run-pipeline-card");
        const overlay = document.getElementById("run-pipeline-modal-overlay");
        card.classList.remove('show');
        overlay.classList.remove('show');

        setTimeout(() => {
            Utils.removeElementById("run-pipeline-card");
            Utils.removeElementById("run-pipeline-modal-overlay");
        }, 300);
    }

    bindEvents() {
        document.getElementById("run-start-pipeline").addEventListener("click", () => this.runPipeline());
        document.getElementById("cancel-run-pipeline").addEventListener("click", () => this.closeModal());
        document.getElementById("run-pipeline-modal-overlay").addEventListener("click", () => this.closeModal());
    }

    runPipeline() {
        const content = this.editor.getValue().trim();
        try {
            fetch(`${baseUrl}${pipelineUrl}/${this.pipelineName}/build`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/yaml',
                },
                body: content
            })
                .then(response => response.json())
                .then(res => {
                    if (res.code === 0) {
                        new TaskModal(res.data.name)
                    } else {
                        alert("流水线运行失败: " + Utils.escapeHTML(res.message));
                    }
                });
        } catch (e) {
            alert("Error: " + e.message);
        }
        this.closeModal();
    }
}

class PipelineModal {
    constructor(pipelineName) {
        this.editor = null;
        this.pipelineName = pipelineName;
        this.webSocketManager = null;
        this.pipeline = null;
        this.currentPage = 1;
        this.rowsPerPage = 15;
        this.totalPage = 0;
        this.tasks = [];
        this.init();
    }

    init() {
// 获取任务详细
        fetch(`${baseUrl}${pipelineUrl}/${this.pipelineName}`)
            .then(response => response.json())
            .then(res => {
                this.pipeline = res.data;
                this.start();
            }).catch(error => {
            console.log('There was a problem with the fetch operation:', error);
            throw error;
        })
    }

    start() {
        this.webSocketManager = new WebSocketManager(`${wsBaseUrl}${pipelineUrl}/${this.pipelineName}/build`, this.updateData.bind(this), (error) => {
            alert('WebSocket error: ' + error);
            this.closeModal();
        })
        this.createModal();
        this.addEventListeners();
    }

    updateData(res) {
        if (res.data && res.data.tasks) {
            this.tasks = res.data.tasks;
            this.totalPage = res.data.page.total;
            this.currentPage = res.data.page.current;
            this.rowsPerPage = res.data.page.size;
            this.renderPipelineTask();
            this.updatePipelineTaskPagination();
            return;
        }

        // 置空表格, 显示无数据, 页码置为0
        this.tasks = [];
        this.totalPage = 1;
        this.renderPipelineTask();
        this.updatePipelineTaskPagination();
    }

    renderPipelineTask() {
        const taskContainer  = document.getElementById("pipeline-task-list");
        taskContainer.innerHTML = '';
        this.tasks.forEach(task => {
            const link = document.createElement('a');
            link.href = "#";
            link.innerText = task.taskName;
            link.style.display = 'inline-table';
            link.style.padding= '3px 6px';
            link.onclick = this.openTaskCard.bind(this, task.taskName);
            taskContainer.appendChild(link);
        });
        document.getElementById("pipeline-task-page-info").textContent = `第${this.currentPage}页__共${this.totalPage}页`;
    }

    openTaskCard(taskName) {
        new TaskModal(taskName)
        this.closeModal();
    }

    createModal() {
        Utils.removeElementById("pipeline-detail-card");
        Utils.removeElementById("pipeline-detail-modal-overlay");

        const overlay = document.createElement('div');
        overlay.setAttribute("id", "pipeline-detail-modal-overlay");
        overlay.className = 'modal-overlay';
        document.body.appendChild(overlay);

        const card = document.createElement('div');
        card.setAttribute("id", "pipeline-detail-card");
        card.className = 'card-one';
        card.innerHTML = `
            <div class="card-header">
                <button id="pipeline-view-detail-bt" class="button-sure">详情</button>
                <button id="pipeline-view-list-bt" class="button-sure">任务列表</button>
                <span class="card-close" id="close-task-card">&times;</span >
            </div>
            <hr>
            <div id="pipeline-view-detail">
                <h5>名称: ${this.pipelineName}</h5>
                <h5>描述: </h5>
                <div class="step-card-output" style="height: 66px">
                    <pre class="step-card-code">${this.pipeline.desc ? this.pipeline.desc : ''}</pre>
                </div>
                <h5>内容: </h5>
                <div class="step-card-output" style="position: absolute;top: 162px; bottom: 6px; right: 8px; left: 8px;height: auto;">
                    <div id="yaml-editor"></div>
                </div>
            </div>
            <div id="pipeline-view-list" style="display: none">
                <div class="card-body">
                    <div id="pipeline-task-pagination" class="pagination" style="position: fixed; right: 6px;display: flex">
                        <div style="margin-right: 6px;">
                            <button id="pipeline-task-prev-page" class="button-sure">上一页</button>
                            <span id="pipeline-task-page-info">第1页__共1页</span>
                            <button id="pipeline-task-next-page" class="button-sure">下一页</button>
                        </div>
                        <div style="display: flex; align-items: center;">
                            <p style="margin-right: 6px;">每页行数</p>
                            <select id="pipeline-task-page-size" class="page-size">
                                <option value="15">15</option>
                                <option value="25">25</option>
                                <option value="35">35</option>
                                <option value="45">45</option>
                                <option value="55">55</option>
                                <option value="65">65</option>
                                <option value="75">75</option>
                                <option value="85">85</option>
                                <option value="95">95</option>
                            </select>
                        </div>
                    </div>
                    <div id="pipeline-task-list" style="margin-top: 30px; width: 100%;height: calc(100% - 30px);overflow-y: auto"></div>
                </div>
            </div>
        `;
        document.body.appendChild(card);
        card.querySelector("#pipeline-view-detail-bt").addEventListener("click", () => this.closePipelineViewList())
        card.querySelector("#pipeline-view-list-bt").addEventListener("click", () => this.showPipelineViewList())

        Utils.InitializeEditor('yaml-editor', this.pipeline.content, true).then(editor => this.editor = editor);
        setTimeout(() => {
            overlay.classList.add('show');
            card.classList.add('show');
        }, 10);
    }

    closePipelineViewList() {
        document.getElementById("pipeline-view-list").style.display = "none";
        document.getElementById("pipeline-view-detail").style.display = "block";
    }

    showPipelineViewList() {
        document.getElementById("pipeline-view-detail").style.display = "none";
        document.getElementById("pipeline-view-list").style.display = "block";
    }

    updatePipelineTaskPagination() {
        document.getElementById("pipeline-task-prev-page").disabled = this.currentPage === 1;
        document.getElementById("pipeline-task-next-page").disabled = this.currentPage === this.totalPage;
    }

    fetchPipelines() {
        const request = {
            page: this.currentPage,
            size: this.rowsPerPage,
        };
        this.webSocketManager.send(request);
    }

    addEventListeners() {
        document.getElementById('close-task-card').addEventListener('click', () => this.closeModal());
        document.getElementById("pipeline-detail-modal-overlay").addEventListener("click", () => this.closeModal());
        window.addEventListener('keydown', this.handleEscapeKey.bind(this));

        document.getElementById("pipeline-task-prev-page").addEventListener("click", () => {
            if (this.currentPage > 1) {
                this.currentPage--;
                this.fetchPipelines();
            }
        });

        document.getElementById("pipeline-task-next-page").addEventListener("click", () => {
            if (this.currentPage < this.totalPage) {
                this.currentPage++;
                this.fetchPipelines();
            }
        });

        document.getElementById("pipeline-task-page-size").addEventListener("change", (event) => {
            this.rowsPerPage = parseInt(event.target.value);
            this.currentPage = 1;
            this.fetchPipelines();
        });
    }

    handleEscapeKey(event) {
        if (event.key === 'Escape') {
            this.closeModal();
        }
    }

    closeModal() {
        if (this.webSocketManager) {
            this.webSocketManager.close();
            this.webSocketManager = null;
        }
        Utils.removeElementById("pipeline-detail-card");
        Utils.removeElementById("pipeline-detail-modal-overlay");
    };
}

class TaskTable {
    constructor() {
        this.webSocketManager  = null;
        this.currentPage = 1;
        this.rowsPerPage = 15;
        this.tasks = [];
        this.totalPage = 0;

        this.init();
    }

    // 初始化 WebSocket
    init() {
        this.webSocketManager = new WebSocketManager(`${wsBaseUrl}${taskUrl}`, this.handleWebSocketTaskData);
        this.setupTaskEventListeners();
    }

    // 处理 WebSocket 返回的数据
    handleWebSocketTaskData = (res) => {
        if (res.data && res.data.tasks) {
            this.tasks = res.data.tasks;
            this.totalPage = res.data.page.total;
            this.currentPage = res.data.page.current;
            this.rowsPerPage = res.data.page.size;
            this.renderTaskTable();
            this.updateTaskPagination();
            return;
        }
        // 置空表格, 显示无数据, 页码置为0
        this.tasks = [];
        this.totalPage = 1;
        this.renderTaskTable();
        this.updateTaskPagination();
    }

    // 通过 WebSocket 请求任务数据
    fetchTasks() {
        const request = {
            page: this.currentPage,
            size: this.rowsPerPage,
        };
        this.webSocketManager.send(request);
    }

    // 动态渲染表格
    renderTaskTable() {
        const tableBody = document.querySelector("#task-table tbody");
        tableBody.innerHTML = "";

        if (!this.tasks || this.tasks.length === 0) {
            const row = document.createElement("tr");
            row.innerHTML = `<td colspan="7"><div style="display:flex;justify-content:center;align-items:center;">暂无数据</div></td>`;
            tableBody.appendChild(row);
            return
        }
        this.tasks.forEach(task => {
            const row = document.createElement("tr");
            const color = Utils.getStatusColor(task.state || 'unknown');
            row.innerHTML = `
                    <td id="task-${task.name+'-name'}">${task.name}</td>
                    <td id="task-${task.name+'-count'}">${task.count}</td>
                    <td id="task-${task.name+'-message'}" class="message" title=""></td>
                    <td id="task-${task.name+'-start'}">${task.time.start ? task.time.start : '---'}</td>
                    <td id="task-${task.name+'-end'}">${task.time.end ? task.time.end : '---'}</td>
                    <td id="task-${task.name+'-state'}"><div style="background-color: ${color}; border-radius: 6px; padding: 5px 10px;">${task.state}</div></td>
                    <td id="task-${task.name+'-actions'}">
                        <div id="task-dropdown" class="dropdown">
                            <button class="dropbtn">Actions</button>
                            <div class="dropdown-content">
                                <a id="detail-task" href="#">详情</a>
                                <a id="dump-task" href="#">导出</a>
                                ${task.state === 'running' ? '<a id="kill-task" href="#">强杀</a>' : ''}
                                <a id="delete-task" href="#">删除</a>
                            </div>
                        </div>
                    </td>
                `;
            tableBody.appendChild(row);
            let msgDocument = document.getElementById('task-'+task.name + '-message');
            if (!msgDocument) {
                msgDocument = document.createElement('div');
                msgDocument.id = task.name + '-message';
            }
            if (task.message) {
                msgDocument.innerText = task.message;
                msgDocument.setAttribute('title', task.message);
                if (task.message.length > 150) {
                    msgDocument.innerText = task.message.substring(0, 150) + '...';
                }
            }
            row.querySelector("#detail-task").addEventListener("click", () => this.showTaskCard(task.name));
            row.querySelector("#dump-task").addEventListener("click", () => this.dumpTask(task));
            if (row.querySelector("#kill-task") !== null) {
                row.querySelector("#kill-task").addEventListener("click", () => Utils.taskManager(task.name, 'kill'));
            }
            if (row.querySelector("#delete-task") !== null) {
                row.querySelector("#delete-task").addEventListener("click", () => this.deleteTask(task));
            }
        });

        document.getElementById("task-page-info").textContent = `第${this.currentPage}页__共${this.totalPage}页`;
    }

    // 更新分页
    updateTaskPagination() {
        document.getElementById("task-prev-page").disabled = this.currentPage === 1;
        document.getElementById("task-next-page").disabled = this.currentPage === this.totalPage;
    }

    // 设置事件监听器
    setupTaskEventListeners() {
        document.getElementById("task-prev-page").addEventListener("click", () => {
            if (this.currentPage > 1) {
                this.currentPage--;
                this.fetchTasks();
            }
        });

        document.getElementById("task-next-page").addEventListener("click", () => {
            if (this.currentPage < this.totalPage) {
                this.currentPage++;
                this.fetchTasks();
            }
        });

        document.getElementById("task-page-size").addEventListener("change", (event) => {
            this.rowsPerPage = parseInt(event.target.value);
            this.currentPage = 1;
            this.fetchTasks();
        });
    };

    deleteTask(task) {
        const confirmed = confirm(`确定要删除任务 "${task.name}"?`);
        if (confirmed) {
            fetch(`${baseUrl}${taskUrl}/${task.name}`, {
                method: 'DELETE',
            })
                .then(response => {
                    if (!response.ok) {
                        throw new Error('Network response was not ok');
                    }
                    return response.json();
                })
                .catch(error => {
                    console.log('There was a problem with the fetch operation:', error);
                    throw error;
                });
        }
    }

    showTaskCard(taskName) {
        new TaskModal(taskName);
    };

    dumpTask(task) {
        fetch(`${baseUrl}${taskUrl}/${task.name}/dump`, {
            method: 'GET',
        }).then(response => {
            if (!response.ok) {
                throw new Error('Network response was not ok');
            }
            return response.json();
        }).then(res => {
            if (res.code !== 0) {
                alert(res.message);
                throw new Error(res.message);
            }
            const blob = new Blob([res.data], { type: 'application/yaml' });
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `${task.name}.yaml`;
            a.click();
            window.URL.revokeObjectURL(url);
        }).catch(error => {
            console.log('There was a problem with the fetch operation:', error);
            throw error;
        })
    };
}

class PipelineTable {
    constructor() {
        this.webSocketManager  = null;
        this.currentPage = 1;
        this.rowsPerPage = 15;
        this.pipelines = [];
        this.totalPage = 0;

        this.init();
    }

    // 初始化 WebSocket
    init() {
        this.webSocketManager = new WebSocketManager(`${wsBaseUrl}${pipelineUrl}`, this.handleWebSocketPipelineData);
        this.setupPipelineEventListeners();
    }

    // 处理 WebSocket 返回的数据
    handleWebSocketPipelineData = (res) => {
        if (res.data && res.data.pipelines) {
            this.pipelines = res.data.pipelines;
            this.totalPage = res.data.page.total;
            this.currentPage = res.data.page.current;
            this.rowsPerPage = res.data.page.size;
            this.renderPipelineTable();
            this.updatePipelinePagination();
            return;
        }
        // 置空表格, 显示无数据, 页码置为0
        this.pipelines = [];
        this.totalPage = 1;
        this.renderPipelineTable();
        this.updatePipelinePagination();
    }

    // 通过 WebSocket 请求任务数据
    fetchPipelines() {
        const request = {
            page: this.currentPage,
            size: this.rowsPerPage,
        };
        this.webSocketManager.send(request);
    }

    // 动态渲染表格
    renderPipelineTable() {
        const tableBody = document.querySelector("#pipeline-table tbody");
        tableBody.innerHTML = "";

        if (!this.pipelines || this.pipelines.length === 0) {
            const row = document.createElement("tr");
            row.innerHTML = `<td colspan="7"><div style="display:flex;justify-content:center;align-items:center;">暂无数据</div></td>`;
            tableBody.appendChild(row);
            return
        }
        this.pipelines.forEach(pipeline => {
            const row = document.createElement("tr");
            row.innerHTML = `
                    <td id="pipeline-${pipeline.name+'-name'}">${pipeline.name}</td>
                    <td id="pipeline-${pipeline.name+'-tplType'}">${pipeline.tplType}</td>
                    <td id="pipeline-${pipeline.name+'-disable'}">${pipeline.disable ? pipeline.disable : '---'}</td>
                    <td id="pipeline-${pipeline.name+'-actions'}">
                        <div id="pipeline-dropdown" class="dropdown">
                            <button class="dropbtn">Actions</button>
                            <div class="dropdown-content">
                                <a id="detail-pipeline" href="#">详情</a>
                                <a id="edit-pipeline" href="#">编辑</a>
                                <a id="run-pipeline" href="#">运行</a>
                                <a id="delete-pipeline" href="#">删除</a>
                            </div>
                        </div>
                    </td>
                `;
            tableBody.appendChild(row);
            let msgDocument = document.getElementById(pipeline.name + '-message');
            if (!msgDocument) {
                msgDocument = document.createElement('div');
                msgDocument.id = pipeline.name + '-message';
            }
            if (pipeline.message) {
                msgDocument.innerText = pipeline.message;
                msgDocument.setAttribute('title', pipeline.message);
            }
            row.querySelector("#detail-pipeline").addEventListener("click", () => this.showPipelineCard(pipeline.name));
            row.querySelector("#run-pipeline").addEventListener("click", () => this.showRunPipeline(pipeline.name));
            row.querySelector("#edit-pipeline").addEventListener("click", () => this.showEditPipeline(pipeline.name));
            row.querySelector("#delete-pipeline").addEventListener("click", () => this.deletePipeline(pipeline));
        });

        document.getElementById("pipeline-page-info").textContent = `第${this.currentPage}页__共${this.totalPage}页`;
    }

    // 更新分页
    updatePipelinePagination() {
        document.getElementById("pipeline-prev-page").disabled = this.currentPage === 1;
        document.getElementById("pipeline-next-page").disabled = this.currentPage === this.totalPage;
    }

    // 设置事件监听器
    setupPipelineEventListeners() {
        document.getElementById("pipeline-prev-page").addEventListener("click", () => {
            if (this.currentPage > 1) {
                this.currentPage--;
                this.fetchPipelines();
            }
        });

        document.getElementById("pipeline-next-page").addEventListener("click", () => {
            if (this.currentPage < this.totalPage) {
                this.currentPage++;
                this.fetchPipelines();
            }
        });

        document.getElementById("pipeline-page-size").addEventListener("change", (event) => {
            this.rowsPerPage = parseInt(event.target.value);
            this.currentPage = 1;
            this.fetchPipelines();
        });
    };

    deletePipeline(pipeline) {
        const confirmed = confirm(`确定要删除流水线 "${pipeline.name}"?`);
        if (confirmed) {
            fetch(`${baseUrl}${pipelineUrl}/${pipeline.name}`, {
                method: 'DELETE',
            })
                .then(response => {
                    if (!response.ok) {
                        throw new Error('Network response was not ok');
                    }
                    return response.json();
                })
                .catch(error => {
                    console.log('There was a problem with the fetch operation:', error);
                    throw error;
                });
        }
    }

    showPipelineCard(pipelineName) {
        new PipelineModal(pipelineName);
    };

    showRunPipeline(pipelineName) {
        new RunPipelineModal(pipelineName);
    }

    showEditPipeline(pipelineName) {
        new PipelineEditCard(pipelineName);
    }
}

class Main {
    constructor() {
        window.addEventListener("resize", () => this.resize());
        this.createMainContent();
        new TaskTable();
        new PipelineTable();
        this.addEventListeners();
        new EventListener();
    }

    resize() {
        history.replaceState(null, '', '/#');
        location.reload();
    }

    createMainContent() {
        // Create the main div
        const mainDiv = document.createElement('div');
        mainDiv.id = 'main';

        // Create the container div and set the inner HTML
        mainDiv.innerHTML = `
            <div id="container" class="container">
                <div class="header">
                    <div id="menu" class="button">
                        <button id="task-list" class="button-sure">任务</button>
                        <button id="pipeline-list" class="button-sure">流水线</button>
                    </div>
                    <div id="options" class="button">
                        <button id="add-task" class="button-sure" style="display: block">添加</button>
                        <button id="add-pipeline" class="button-sure" style="display: none">添加</button>
                        <p id="title" style="font-size: 15px; margin: auto">任务</p>
                        <p style="margin: auto; font-size: 15px">|</p>
                        
                        <div id="task-table-pagination" class="pagination" style="display: flex">
                            <div style="margin-right: 6px;">
                                <button id="task-prev-page" class="button-sure">上一页</button>
                                <span id="task-page-info">第1页__共1页</span>
                                <button id="task-next-page" class="button-sure">下一页</button>
                            </div>
                            <div style="display: flex; align-items: center;">
                                <p style="margin-right: 6px;">每页行数</p>
                                <select id="task-page-size" class="page-size">
                                    <option value="15">15</option>
                                    <option value="25">25</option>
                                    <option value="35">35</option>
                                    <option value="45">45</option>
                                    <option value="55">55</option>
                                    <option value="65">65</option>
                                    <option value="75">75</option>
                                    <option value="85">85</option>
                                    <option value="95">95</option>
                                </select>
                            </div>
                        </div>
                        
                        <div id="pipeline-table-pagination" class="pagination" style="display: none">
                            <div style="margin-right: 6px;">
                                <button id="pipeline-prev-page" class="button-sure">上一页</button>
                                <span id="pipeline-page-info">第1页__共1页</span>
                                <button id="pipeline-next-page" class="button-sure">下一页</button>
                            </div>
                            <div style="display: flex; align-items: center;">
                                <p style="margin-right: 6px;">每页行数</p>
                                <select id="pipeline-page-size" class="page-size">
                                    <option value="15">15</option>
                                    <option value="25">25</option>
                                    <option value="35">35</option>
                                    <option value="45">45</option>
                                    <option value="55">55</option>
                                    <option value="65">65</option>
                                    <option value="75">75</option>
                                    <option value="85">85</option>
                                    <option value="95">95</option>
                                </select>
                            </div>
                        </div>
                    </div>
                </div>
                <div id="task-table-container" class="table-container" style="display: block">
                    <table id="task-table" class="common-table">
                        <thead style="z-index: 1; position: sticky; top: 0">
                            <tr>
                                <th style="width: 160px">名称</th>
                                <th style="width: 48px;">步骤数</th>
                                <th>消息</th>
                                <th style="width: 180px;">开始时间</th>
                                <th style="width: 180px;">结束时间</th>
                                <th style="width: 48px;">状态</th>
                                <th style="width: 48px;">动作</th>
                            </tr>
                        </thead>
                        <tbody>
                            <!-- Rows will be dynamically inserted here -->
                        </tbody>
                    </table>
                </div>
                <div id="pipeline-table-container" class="table-container" style="display: none">
                    <table id="pipeline-table" class="common-table">
                        <thead>
                            <tr>
                                <th>名称</th>
                                <th>模板类型</th>
                                <th>禁用</th>
                                <th style="width: 48px;">动作</th>
                            </tr>
                        </thead>
                        <tbody>
                            <!-- Rows will be dynamically inserted here -->
                        </tbody>
                    </table>
                </div>
            </div>
        `;

        // Append the newly created main div to the body
        document.body.appendChild(mainDiv);
        mainDiv.querySelector("#task-list").addEventListener("click", () => this.closePipelineTable());
        mainDiv.querySelector("#pipeline-list").addEventListener("click", () => this.showPipelineTable());
    }

    addEventListeners() {
        document.getElementById("add-task").addEventListener("click", () => new TaskAddCard());
        document.getElementById("add-pipeline").addEventListener("click", () => new PipelineAddCard());
    }

    closePipelineTable() {
        document.getElementById("title").innerText = "任务";
        document.getElementById("task-table-container").style.display = "block";
        document.getElementById("add-task").style.display = "block";
        document.getElementById("task-table-pagination").style.display = "flex";
        document.getElementById("pipeline-table-container").style.display = "none";
        document.getElementById("add-pipeline").style.display = "none";
        document.getElementById("pipeline-table-pagination").style.display = "none";
    }

    showPipelineTable() {
        document.getElementById("title").innerText = "流水线";
        document.getElementById("task-table-container").style.display = "none";
        document.getElementById("add-task").style.display = "none";
        document.getElementById("task-table-pagination").style.display = "none";
        document.getElementById("pipeline-table-container").style.display = "block";
        document.getElementById("add-pipeline").style.display = "block";
        document.getElementById("pipeline-table-pagination").style.display = "flex";
    }
}

class EventListener {
    constructor() {
        this.initEventContainer();
        this.listenForEvents();
    }

    // 初始化事件容器
    initEventContainer() {
        const eventContainer = document.createElement('div');
        eventContainer.id = 'event-container';
        eventContainer.className = 'event-container';
        document.body.appendChild(eventContainer);
    }

    // 开始监听事件源
    listenForEvents() {
        const eventSource = new EventSource(eventUrl);

        eventSource.onmessage = (event) => {
            this.displayEventMessage(event.data);
        };

        eventSource.onerror = (event) => {
            console.error('SSE connection error:', event);
        };
    }

    // 展示事件信息
    displayEventMessage(message) {
        const eventContainer = document.getElementById('event-container');
        const messageElement = document.createElement('p');
        messageElement.textContent = message;

        eventContainer.appendChild(messageElement);

        // 保留最后9条消息，删除最早的
        if (eventContainer.children.length > 9) {
            eventContainer.removeChild(eventContainer.firstChild);
        }
    }
}

