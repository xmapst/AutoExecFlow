let highestZIndex = 1000;

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
                    width: 250,
                    height: 80,
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
                    width: 250,
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
                    img: '/img/node-icon.png',
                },
                name: 'node-icon',
            });
            group.addShape('text', {
                attrs: {
                    textBaseline: 'top',
                    y: 2,
                    x: 24,
                    lineHeight: 20,
                    text: cfg.detail.name,
                    fill: '#fff',
                },
                name: 'title-text',
            });
            if (cfg.detail.count !== undefined) {
                group.addShape('text', {
                    attrs: {
                        textBaseline: 'top',
                        y: 25,
                        x: 6,
                        lineHeight: 3,
                        text: "步骤数: " + cfg.detail.count,
                        fill: 'rgba(0,0,0, 0.4)',
                    },
                    name: 'title-count',
                });
            }
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
                    x: 120,
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
                    text: "开始时间: " + cfg.detail.time.start,
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
                    text: "结束时间: " + cfg.detail.time.end,
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
        })
        this.createModal();
        this.addEventListeners();
    };

    createModal() {
        Utils.removeElementById("task-card");

        const card = document.createElement('div');
        card.setAttribute("id", "task-card");
        card.className = 'task-card';
        card.innerHTML = `
                <div class="card-header">
                    <h3 style="margin-left: 15px;">名称: ${this.task.name}</h3>
                    <span class="card-close" id="closeTaskCard">&times;</span >
                </div>
                <hr>
                <div id="task-card-body" class="card-body"></div>
            `;

        document.body.appendChild(card);
    };

    createGraph(data) {
        const card = document.getElementById('task-card');
        const width = card.scrollWidth - 30;
        const height = card.scrollHeight -96 || 500;
        this.graph  = new G6.Graph({
            container: "task-card-body",
            height: height,
            width: width,
            layout: {
                type: 'dagre',
                rankdir: 'LR',
                align: 'UL',
                nodesep: 30,
                ranksep: 120,
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
                };
                const item = this.graph.findById(node.id);
                if (item) {
                    item.detail = node.detail;
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
        document.getElementById('closeTaskCard').addEventListener('click', () => this.closeModal());
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

        Utils.removeElementById('task-card');
        this.closeAllStepModals();
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
        this.WebSocketManager = new WebSocketManager(`${wsBaseUrl}${taskUrl}/${this.taskName}/step/${this.step.name}`,this.updateStepOutput, ()=> {
            const outputElement = document.getElementById('step-output-text');
            outputElement.innerHTML = `<pre class="step-card-code">${Utils.escapeHTML(this.step.msg)}</pre>`;
        });
        this.createModal();
        this.addEventListeners();
    };

    createModal() {
        const existingCard = document.getElementById(this.step.name + "-step-card");
        if (existingCard) {
            existingCard.style.zIndex = ++highestZIndex;
            return;
        }

        const card = document.createElement('div');
        card.setAttribute("id", this.step.name + "-step-card");
        card.className = 'step-card';
        card.style.position = 'absolute';
        card.style.top = '100px';
        card.style.left = '100px';
        card.style.zIndex = ++highestZIndex;
        card.innerHTML = `
                <div id="${this.step.name + '-step-card-header'}" class="card-header">
                    <h3>名称: ${this.step.name}</h3>
                    <h5>类型: ${this.step.type}</h5>
                    <span class="card-close" id="${this.step.name + '-closeStepCard'}">&times;</span>
                </div>
                <hr>
                <div id="${this.step.name + '-step-card-body'}" class="card-body">
                    <h5>输入: </h5>
                    <div id="${this.step.name + '-step-input'}" class="step-card-input">
                        <pre id="${this.step.name + '-step-input-text'}" class="step-card-code"></pre>
                    </div>
                    <h5>输出: </h5>
                    <div id="${this.step.name + '-step-output'}" class="step-card-output">
                        <pre id="${this.step.name + '-step-output-text'}" class="step-card-code"></pre>
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
        document.getElementById(this.step.name + '-closeStepCard').addEventListener('click', () => this.closeModal());
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
        this.show()
    }

    show() {
        Utils.removeElementById("add-task-card");

        const card = document.createElement('div');
        card.setAttribute("id", "add-task-card");
        card.className = 'task-card';
        card.innerHTML = `
            <div class="card-header">
                <h3 style="margin-left: 15px;">添加任务</h3>
                <span class="card-close" onclick="Utils.removeElementById('add-task-card')">&times;</span >
            </div>
            <div class="create-content">
                <textarea id="yamlEditor"></textarea>
                <div class="create-footer">
                    <button id="createTask" class="btn-create">创建</button>
                    <button id="cancelModal" class="btn-cancel">取消</button>
                </div>
            </div>
        `;
        document.body.appendChild(card);

        this.initializeEditor();
        this.bindEvents();
    }

    initializeEditor() {
        const yamlEditorElement = document.getElementById("yamlEditor");
        this.editor = CodeMirror.fromTextArea(yamlEditorElement, {
            mode: "yaml",
            theme: "material-darker",
            lineNumbers: true,
        });
    }

    bindEvents() {
        document.getElementById("createTask").addEventListener("click", () => this.createTask());
        document.getElementById("cancelModal").addEventListener("click", () => Utils.removeElementById("add-task-card"));

        document.getElementById("yamlEditor").addEventListener("keydown", (event) => this.handleTab(event));

        window.onkeydown = (event) => {
            if (event.key === 'Escape') {
                Utils.removeElementById("add-task-card");
            }
        };
    }

    handleTab(event) {
        if (event.key === 'Tab') {
            event.preventDefault();
            const cursor = this.editor.getCursor();
            const line = this.editor.getLine(cursor.line);
            const before = line.slice(0, cursor.ch);
            const after = line.slice(cursor.ch);
            this.editor.replaceRange('\t' + after, { line: cursor.line, ch: 0 });
            this.editor.setCursor({ line: cursor.line, ch: before });
        }
    }

    createTask() {
        const yamlContent = this.editor.getValue();
        if (yamlContent === "") {
            alert("请输入YAML内容");
            return;
        }
        try {
            jsyaml.load(yamlContent);
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
                        alert("任务创建成功");
                        Utils.removeElementById("add-task-card");
                    } else {
                        alert("任务创建失败: " + Utils.escapeHTML(data.msg));
                    }
                });
        } catch (e) {
            alert("Error: " + e.message);
            Utils.removeElementById("add-task-card");
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
        if (res.data.tasks) {
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
        const tableBody = document.querySelector("#taskTable tbody");
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
                    <td id="${task.name+'-message'}"></td>
                    <td id="${task.name+'-state'}"><div style="background-color: ${color}; border-radius: 6px; padding: 6px; display: flex; justify-content: center; align-items: center;">${task.state}</div></td>
                    <td id="${task.name+'-actions'}">
                        <div class="dropdown">
                            <button class="dropbtn">Actions</button>
                            <div class="dropdown-content">
                                <a href="#" id="detailTask">详情</a>
                                <a href="#" id="killTask">强杀</a>
                                <a href="#" id="deleteTask">删除</a>
                            </div>
                        </div>
                    </td>
                `;
                tableBody.appendChild(row);
                const msgDocument = document.getElementById(task.name + '-message');
                if (task.msg) {
                    msgDocument.innerText = task.msg;
                }
                row.querySelector("#detailTask").addEventListener("click", () => this.showTaskCard(task));
                row.querySelector("#killTask").addEventListener("click", () => this.killTask(task));
                row.querySelector("#deleteTask").addEventListener("click", () => this.deleteTask(task));
            });
        }

        document.getElementById("pageInfo").textContent = `第${this.currentPage}页__共${this.totalPage}页`;
    }

    // 更新分页
    updatePagination() {
        document.getElementById("prevPage").disabled = this.currentPage === 1;
        document.getElementById("nextPage").disabled = this.currentPage === this.totalPage;
    }

    // 设置事件监听器
    setupEventListeners() {
        document.getElementById("prevPage").addEventListener("click", () => {
            if (this.currentPage > 1) {
                this.currentPage--;
                this.fetchTasks();
            }
        });

        document.getElementById("nextPage").addEventListener("click", () => {
            if (this.currentPage < this.totalPage) {
                this.currentPage++;
                this.fetchTasks();
            }
        });

        document.getElementById("pageSize").addEventListener("change", (event) => {
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

    killTask(task) {
        if (task.state !== 'running') {
            alert('任务已结束，无法杀死');
            return;
        }
        const confirmed = confirm(`确定要杀死任务 "${task.name}"?`);
        if (confirmed) {
            fetch(`${baseUrl}${taskUrl}/${task.name}?action=kill`, {
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
    };

    showTaskCard(task) {
        new TaskModal(task);
    };
}

class Main {
    constructor() {
        this.createMainContent();
        this.taskTable = new TaskTable();
        this.addEventListeners();
    }

    createMainContent() {
        // Create the main div
        const mainDiv = document.createElement('div');
        mainDiv.id = 'main';

        // Create the container div and set the inner HTML
        mainDiv.innerHTML = `
            <div id="container" class="container">
                <div class="header">
                    <button id="addTask" class="dropbtn" style="margin-right: 20px;">添加</button>
                </div>
                <table id="taskTable">
                    <thead>
                        <tr>
                            <th>名称</th>
                            <th style="width: 48px;">步骤数</th>
                            <th>消息</th>
                            <th style="width: 48px;">状态</th>
                            <th style="width: 48px;">动作</th>
                        </tr>
                    </thead>
                    <tbody>
                        <!-- Rows will be dynamically inserted here -->
                    </tbody>
                </table>
                <div class="pagination">
                    <div style="margin-right: 20px;">
                        <button id="prevPage" class="dropbtn">上一页</button>
                        <span id="pageInfo">第1页__共1页</span>
                        <button id="nextPage" class="dropbtn">下一页</button>
                    </div>
                    <div style="display: flex; align-items: center;">
                        <p>每页行数</p>
                        <select id="pageSize">
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
        document.getElementById("addTask").addEventListener("click", () => new TaskAddCard());

        window.addEventListener("resize", () => location.reload());
    }
}

