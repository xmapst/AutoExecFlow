let highestZIndex = 1000;
require.config({ paths: { 'vs': baseUrl+basePath+'/vs' } });
const taskTpl = `# 异步执行, 可选, 默认并行,自定义编排时需要设置为true
Async: true
# 超时时间, 可选, 默认48小时
Timeout: 2m
# 任务名称, 可选, 默认自动生成
Name: 测试
# 全局环境变量, 可选, key: value形式
Env:
  Test: "test_env"
# 步骤列表, 不能为空
Step:
    # 步骤名称, 唯一, 可选[当自定义编排是必须设置], 默认自动生成
  - Name: 步骤2
    # 超时时间, 可选, 默认任务级超时时间
    Timeout: 2m
    # 依赖步骤, 可选[自定义编排时用到]
    Depends:
      - 步骤1
    # 局部环境变量, 会覆盖同名的全局变量, 可选, key: value形式
    Env:
      Test: "test_env"
    # 类型
    Type: sh
    # 内容
    Content: |-
      ping 1.1.1.1
  - Name: 步骤1
    Timeout: 2m
    Env:
      Test: "test_env"
    Type: sh
    Content: |-
      ping 1.1.1.1
`

G6.registerNode(
    'custom-node',
    {
        drawShape: function drawShape(cfg, group) {
            const color = Utils.getStatusColor(cfg.detail.state || 'unknown');
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
                    text: cfg.detail.name,
                    fill: '#fff',
                },
                name: 'title-text',
            });
            if (cfg.detail.code !== undefined) {
                group.addShape('text', {
                    attrs: {
                        textBaseline: 'top',
                        y: 25,
                        x: 6,
                        lineHeight: 3,
                        text: "状态码: " + cfg.detail.code,
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
                    text: "状态: " + cfg.detail.state,
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
                    text: cfg.detail.time.start ? "开始时间: " + cfg.detail.time.start : '开始时间: ---',
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
                    text: cfg.detail.time.end ? "结束时间: " + cfg.detail.time.end : '结束时间: ---',
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
            'running': '#00FFFF',
            'stopped': '#2FC25B',
            'failed': '#F4664A',
            'pending': '#F6BD16',
            'timeout': '#FACC14',
            'canceled': '#B37FD3',
            'skipped': '#1890FF',
            'unknown': '#DCDCDC'
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
}

// 通用的WebSocket管理器
class WebSocketManager {
    constructor(url, onMessage, onError) {
        this.socket = new WebSocket(url);
        this.socket.onopen = () => console.log("WebSocket connection established.");
        this.socket.onmessage = (event) => {
            const data = JSON.parse(event.data)
            onMessage(data)
        };
        this.socket.onerror = (error) => {
            console.error('WebSocket error:', error);
            if (onError) onError(error);
        };
        this.socket.onclose = () => console.log("WebSocket connection closed.");
    }

    send(data) {
        if (this.socket.readyState === WebSocket.OPEN) {
            this.socket.send(JSON.stringify(data));
        }
    }

    close() {
        if (this.socket) {
            this.socket.close(1000, "Normal closure");
        }
    }
}


class TaskModal {
    constructor(task) {
        this.WebSocketManager = null;
        this.task = task;
        this.graph = null;
        this.stepModals = [];
        this.init();
    };

    init() {
        this.WebSocketManager = new WebSocketManager(`${wsBaseUrl}${taskUrl}/${this.task.name}`, this.updateGraphData, (error) => {
            alert('WebSocket error: ' + error);
            this.closeModal();
        })
        this.createModal();
        this.addEventListeners();
    };

    createModal() {
        Utils.removeElementById('task-card');
        Utils.removeElementById("modal-overlay");

        const overlay = document.createElement('div');
        overlay.setAttribute("id", "modal-overlay");
        overlay.className = 'modal-overlay';
        document.body.appendChild(overlay);

        const card = document.createElement('div');
        card.setAttribute("id", "task-card");
        card.className = 'task-card';
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
                             <pre class="env">${Object.entries(this.task.env).map(([key, value]) => `${key}=${value}`).join('\n')}</pre>
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
                const task = e.item.getModel().task;
                const step = e.item.getModel().detail;
                // 判断节点是否找到
                if (!step || task.state !== 'running' ) {
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
                const task = item.getModel().task;
                const step = item.getModel().detail;
                // 根据点击的菜单项执行相应的操作
                if (target.id === 'kill-step') {
                    Utils.stepManager(task.name, step.name, 'kill')
                } else if (target.id === 'pause-step') {
                    Utils.stepManager(task.name, step.name, 'pause')
                } else if (target.id === 'resume-step') {
                    Utils.stepManager(task.name, step.name, 'resume')
                }
            },
        });
        const toolbar = new G6.ToolBar({
            getContent: () => {
                return `<ul class="g6-component-toolbar" style="top: 0px; left: 1011px;">
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
            const model = evt.item.getModel();
            this.openStepModal(model.detail);
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
                    detail: step,
                    task: this.task,
                };
                const item = this.graph.findById(node.id);
                if (item) {
                    item.detail = node.detail;
                    item.task = this.task;
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
                detail: step,
                task: this.task,
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

    openStepModal(step) {
        const stepModal = new StepModal(this.task.name, step);
        this.stepModals.push(stepModal);
    }

    addEventListeners() {
        document.getElementById('close-task-card').addEventListener('click', () => this.closeModal());
        document.getElementById("modal-overlay").addEventListener("click", () => this.closeModal());
        window.addEventListener('keydown', this.handleEscapeKey.bind(this));
    };

    handleEscapeKey(event) {
        if (event.key === 'Escape') {
            this.closeModal();
        }
    }

    closeModal() {
        if (this.WebSocketManager) {
            this.WebSocketManager.close();
            this.WebSocketManager = null;
        }
        this.closeAllStepModals();
        const card = document.getElementById("task-card");
        const overlay = document.getElementById("modal-overlay");
        card.classList.remove('show');
        overlay.classList.remove('show');
        setTimeout(() => {
            Utils.removeElementById("task-card");
            Utils.removeElementById("modal-overlay");
        }, 300);
        window.removeEventListener('keydown', this.handleEscapeKey.bind(this));
    };
}

class StepModal {
    constructor(taskName, step) {
        this.taskName = taskName;
        this.WebSocketManager = null;
        this.step = step;
        this.isDragging = false;
        this.offsetX = 0;
        this.offsetY = 0;
        this.init();
    };

    init() {
        const existingCard = document.getElementById(this.step.name + "-step-card");
        if (existingCard) {
            existingCard.style.zIndex = ++highestZIndex;
            return;
        }
        this.WebSocketManager = new WebSocketManager(`${wsBaseUrl}${taskUrl}/${this.taskName}/step/${this.step.name}`,this.updateStepOutput, ()=> {
            const outputElement = document.getElementById('step-output-text');
            outputElement.innerHTML = `<pre class="step-card-code">${Utils.escapeHTML(this.step.msg)}</pre>`;
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
                                <pre class="env">${Object.entries(this.step.env).map(([key, value]) => `${key}=${value}`).join('\n')}</pre>
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
        if (this.WebSocketManager) {
            this.WebSocketManager.close();
            this.WebSocketManager = null;
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
        this.show();
    }

    show() {
        Utils.removeElementById("add-task-card");
        Utils.removeElementById("modal-overlay");

        const overlay = document.createElement('div');
        overlay.setAttribute("id", "modal-overlay");
        overlay.className = 'modal-overlay';
        document.body.appendChild(overlay);

        const card = document.createElement('div');
        card.setAttribute("id", "add-task-card");
        card.className = 'task-card';
        card.innerHTML = `
            <div class="card-header">
                <div class="button" style="position: fixed; top: 6px; right: 12px;">
                    <button id="create-task" class="button-sure">创建</button>
                    <button id="cancel-modal" class="button-cancel">取消</button>
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

        this.initializeEditor();
        this.bindEvents();
    }

    initializeEditor() {
        // 初始化 Monaco Editor
        require(['vs/editor/editor.main'], () => {
            this.editor = monaco.editor.create(document.getElementById('yaml-editor'), {
                value: taskTpl,
                language: 'yaml',
                theme: 'vs-dark',
                autoIndent: true,
                automaticLayout: true,
                overviewRulerBorder: false,
                foldingStrategy: 'indentation',
                lineNumbers: 'on',
                minimap: { enabled: false },
                tabSize: 2,
                mouseWheelZoom: true,
                formatOnType: true,
                formatOnPaste: true,
                cursorStyle: 'line',
                fontSize: 12,
            });
        });
    }

    closeModal() {
        const card = document.getElementById("add-task-card");
        const overlay = document.getElementById("modal-overlay");
        card.classList.remove('show');
        overlay.classList.remove('show');

        setTimeout(() => {
            Utils.removeElementById("add-task-card");
            Utils.removeElementById("modal-overlay");
        }, 300);
    }

    bindEvents() {
        document.getElementById("create-task").addEventListener("click", () => this.createTask());
        document.getElementById("cancel-modal").addEventListener("click", () => this.closeModal());
        document.getElementById("modal-overlay").addEventListener("click", () => this.closeModal());
    }

    createTask() {
        const yamlContent = this.editor.getValue();
        if (yamlContent === "") {
            alert("请输入YAML内容");
            return;
        }
        try {
            fetch(`${baseUrl}${taskV2Url}`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/yaml',
                },
                body: yamlContent,
            })
                .then(response => response.json())
                .then(data => {
                    if (data.code === 0) {
                        alert("任务添加成功");
                    } else {
                        alert("任务添加失败: " + Utils.escapeHTML(data.msg));
                    }
                });
        } catch (e) {
            alert("Error: " + e.message);
            this.closeModal();
        }
    }
}

class TaskTable {
    constructor() {
        this.webSocketManager  = null;
        this.currentPage = 1;
        this.rowsPerPage = 10;
        this.tasks = [];
        this.totalPage = 0;

        this.init();
    }

    // 初始化 WebSocket
    init() {
        this.webSocketManager = new WebSocketManager(`${wsBaseUrl}${taskUrl}?page=${this.currentPage}&size=${this.rowsPerPage}`, this.handleWebSocketData);
        this.setupEventListeners();
    }

    // 处理 WebSocket 返回的数据
    handleWebSocketData = (res) => {
        if (res.data && res.data.tasks) {
            this.tasks = res.data.tasks;
            this.totalPage = res.data.page.total;
            this.currentPage = res.data.page.current;
            this.rowsPerPage = res.data.page.size;
            this.renderTable();
            this.updatePagination();
        }
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
    renderTable() {
        const tableBody = document.querySelector("#task-table tbody");
        tableBody.innerHTML = "";

        if (!this.tasks || this.tasks.length === 0) {
            const row = document.createElement("tr");
            row.innerHTML = `<td colspan="5"><div style="display:flex;justify-content:center;align-items:center;">暂无数据</div></td>`;
            tableBody.appendChild(row);
        } else {
            this.tasks.forEach(task => {
                const row = document.createElement("tr");
                const color = Utils.getStatusColor(task.state || 'unknown');
                row.innerHTML = `
                    <td id="${task.name+'-name'}">${task.name}</td>
                    <td id="${task.name+'-count'}">${task.count}</td>
                    <td id="${task.name+'-message'}" class="message" title=""></td>
                    <td id="${task.name+'-start'}">${task.time.start ? task.time.start : '---'}</td>
                    <td id="${task.name+'-end'}">${task.time.end ? task.time.end : '---'}</td>
                    <td id="${task.name+'-state'}"><div style="background-color: ${color}; border-radius: 6px; padding: 5px 10px;">${task.state}</div></td>
                    <td id="${task.name+'-actions'}">
                        <div class="dropdown">
                            <button class="dropbtn">Actions</button>
                            <div class="dropdown-content">
                                <a href="#" id="detail-task">详情</a>
                                ${task.state === 'running' ? '<a href="#" id="kill-task">强杀</a>' : ''}
                                ${task.state === 'stopped' || task.state === 'failed' ? '<a href="#" id="delete-task">删除</a>' : ''}
                            </div>
                        </div>
                    </td>
                `;
                tableBody.appendChild(row);
                const msgDocument = document.getElementById(task.name + '-message');
                if (task.msg) {
                    msgDocument.innerText = task.msg;
                    msgDocument.setAttribute('title', task.msg);
                }
                row.querySelector("#detail-task").addEventListener("click", () => this.showTaskCard(task));
                if (row.querySelector("#kill-task") !== null) {
                    row.querySelector("#kill-task").addEventListener("click", () => Utils.taskManager(task.name, 'kill'));
                }
                if (row.querySelector("#delete-task") !== null) {
                    row.querySelector("#delete-task").addEventListener("click", () => this.deleteTask(task));
                }
            });
        }

        document.getElementById("page-info").textContent = `第${this.currentPage}页__共${this.totalPage}页`;
    }

    // 更新分页
    updatePagination() {
        document.getElementById("prev-page").disabled = this.currentPage === 1;
        document.getElementById("next-page").disabled = this.currentPage === this.totalPage;
    }

    // 设置事件监听器
    setupEventListeners() {
        document.getElementById("prev-page").addEventListener("click", () => {
            if (this.currentPage > 1) {
                this.currentPage--;
                this.fetchTasks();
            }
        });

        document.getElementById("next-page").addEventListener("click", () => {
            if (this.currentPage < this.totalPage) {
                this.currentPage++;
                this.fetchTasks();
            }
        });

        document.getElementById("page-size").addEventListener("change", (event) => {
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

    showTaskCard(task) {
        new TaskModal(task);
    };
}

class Main {
    constructor() {
        this.createMainContent();
        new TaskTable();
        this.addEventListeners();
        new EventListener();
    }

    createMainContent() {
        // Create the main div
        const mainDiv = document.createElement('div');
        mainDiv.id = 'main';

        // Create the container div and set the inner HTML
        mainDiv.innerHTML = `
            <div id="container" class="container">
                <div class="header">
                    <div class="button">
                        <button id="add-task" class="button-sure">添加</button>
                    </div>
                </div>
                <div class="table-container">
                    <table id="task-table">
                        <thead>
                            <tr>
                                <th>名称</th>
                                <th style="width: 48px;">步骤数</th>
                                <th>消息</th>
                                <th style="width: 162px;">开始时间</th>
                                <th style="width: 162px;">结束时间</th>
                                <th style="width: 48px;">状态</th>
                                <th style="width: 48px;">动作</th>
                            </tr>
                        </thead>
                        <tbody>
                            <!-- Rows will be dynamically inserted here -->
                        </tbody>
                    </table>
                </div>
                <div class="pagination">
                    <div style="margin-right: 6px;">
                        <button id="prev-page" class="button-sure">上一页</button>
                        <span id="page-info">第1页__共1页</span>
                        <button id="next-page" class="button-sure">下一页</button>
                    </div>
                    <div style="display: flex; align-items: center;">
                        <p style="margin-right: 6px;">每页行数</p>
                        <select id="page-size">
                            <option value="10">10</option>
                            <option value="15">15</option>
                            <option value="20">20</option>
                            <option value="25">25</option>
                            <option value="30">30</option>
                            <option value="35">35</option>
                        </select>
                    </div>
                </div>
            </div>
        `;

        // Append the newly created main div to the body
        document.body.appendChild(mainDiv);
    }

    addEventListeners() {
        document.getElementById("add-task").addEventListener("click", () => new TaskAddCard());

        window.addEventListener("resize", () => location.reload());
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

        // 保留最后三条消息，删除最早的
        if (eventContainer.children.length > 9) {
            eventContainer.removeChild(eventContainer.firstChild);
        }
    }
}

