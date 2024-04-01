<template>
  <div>
    <a-card :bordered="false">
      <a-row :gutter="24">
        <a-col :sm="6" :xs="24">
          <info :title="$t('pool.worker')" :value="pool.size" :bordered="true" />
        </a-col>
        <a-col :sm="6" :xs="24">
          <info :title="$t('pool.running')" :value="pool.running" :bordered="true" />
        </a-col>
        <a-col :sm="6" :xs="24">
          <info :title="$t('pool.waiting')" :value="pool.waiting" :bordered="true" />
        </a-col>
        <a-col :sm="6" :xs="24">
          <info :title="$t('pool.total')" :value="pool.total"/>
        </a-col>
      </a-row>
    </a-card>


  </div>
</template>

<script>
import { onUnmounted, ref } from 'vue'
import { getPoolState } from '@/api/pool'
import { Radar } from '@/components'
import Info from './components/Info'

const timer = ref(null)

export default {
  name: 'Workplace',
  components: {
    Radar,
    Info
  },
  data () {
    return {
      pool: {
        size: 0,
        total: 0,
        running: 0,
        waiting: 0
      }
    }
  },
  computed: {
  },
  created () {
    this.reload()
  },
  mounted () {
    timer.value = setInterval(() => {
      this.reload()
    }, 1000 * 3)

    onUnmounted(() => {
      clearInterval(timer.value)
      timer.value = null
    })
  },
  methods: {
    async reload() {
      await getPoolState().then(res => {
        this.pool = res.data
      }).catch(err => {
        this.$message.error(err.message)
      })
    }
  }
}
</script>

<style lang="less" scoped>
@import './Index.less';

.system_state {
  padding: 10px;
}

.card_item {
  height: 280px;
}

.project-list {
  .card-title {
    font-size: 0;

    a {
      color: rgba(0, 0, 0, 0.85);
      margin-left: 12px;
      line-height: 24px;
      height: 24px;
      display: inline-block;
      vertical-align: top;
      font-size: 14px;

      &:hover {
        color: #1890ff;
      }
    }
  }

  .card-description {
    color: rgba(0, 0, 0, 0.45);
    height: 44px;
    line-height: 22px;
    overflow: hidden;
  }

  .project-item {
    display: flex;
    margin-top: 8px;
    overflow: hidden;
    font-size: 12px;
    height: 20px;
    line-height: 20px;

    a {
      color: rgba(0, 0, 0, 0.45);
      display: inline-block;
      flex: 1 1 0;

      &:hover {
        color: #1890ff;
      }
    }

    .datetime {
      color: rgba(0, 0, 0, 0.25);
      flex: 0 0 auto;
      float: right;
    }
  }

  .ant-card-meta-description {
    color: rgba(0, 0, 0, 0.45);
    height: 44px;
    line-height: 22px;
    overflow: hidden;
  }
}

.item-group {
  padding: 20px 0 8px 24px;
  font-size: 0;

  a {
    color: rgba(0, 0, 0, 0.65);
    display: inline-block;
    font-size: 14px;
    margin-bottom: 13px;
    width: 25%;
  }
}

.members {
  a {
    display: block;
    margin: 12px 0;
    line-height: 24px;
    height: 24px;

    .member {
      font-size: 14px;
      color: rgba(0, 0, 0, 0.65);
      line-height: 24px;
      max-width: 100px;
      vertical-align: top;
      margin-left: 12px;
      transition: all 0.3s;
      display: inline-block;
    }

    &:hover {
      span {
        color: #1890ff;
      }
    }
  }
}

.mobile {
  .project-list {
    .project-card-grid {
      width: 100%;
    }
  }

  .more-info {
    border: 0;
    padding-top: 16px;
    margin: 16px 0 16px;
  }

  .headerContent .title .welcome-text {
    display: none;
  }
}
</style>
