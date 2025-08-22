<template>
  <div class="desk-title-bar" :class="theme">
    <div v-if="$slots.icon" class="left-bar">
      <slot name="icon"></slot>
    </div>

    <div class="title">
      <slot></slot>
      {{ title }}
    </div>

    <div class="controls">
      <a v-if="showMinimize" @click="onMinimizeClick" title="最小化" class="button minimize">
        <svg class="icon" style="transform: scaleY(0.5)"><use xlink:href="#desk-icon-minimize"></use></svg>
      </a>

      <a v-if="showMaximize" @click="onMaxIconClick" :title="isMaximized_inner ? '还原' : '最大化'" class="button maximize">
        <svg v-if="isMaximized_inner" class="icon"><use xlink:href="#desk-icon-maximize"></use></svg>
        <svg v-if="!isMaximized_inner" class="icon"><use xlink:href="#desk-icon-restore"></use></svg>
      </a>

      <a v-if="showClose" @click="onCloseClick" title="关闭" class="button close">
        <svg class="icon"><use xlink:href="#desk-icon-close"></use></svg>
      </a>
    </div>

    <svg style="position: absolute; width: 0; height: 0; overflow: hidden">
      <!--最小化-->
      <symbol id="desk-icon-minimize" x="0px" y="0px" viewBox="0 0 10.2 1">
        <rect width="10.2" height="1" />
      </symbol>

      <!--最大化-->
      <symbol id="desk-icon-maximize" x="0px" y="0px" viewBox="0 0 10.2 10.2">
        <path d="M2.1,0v2H0v8.1h8.2v-2h2V0H2.1z M7.2,9.2H1.1V3h6.1V9.2z M9.2,7.1h-1V2H3.1V1h6.1V7.1z" />
      </symbol>

      <!--还原-->
      <svg id="desk-icon-restore" x="0px" y="0px" viewBox="0 0 10.2 10.1">
        <path d="M0,0v10.1h10.2V0H0z M9.2,9.2H1.1V1h8.1V9.2z" />
      </svg>

      <!--关闭-->
      <symbol id="desk-icon-close" x="0px" y="0px" viewBox="0 0 10.2 10.2">
        <polygon points="10.2,0.7 9.5,0 5.1,4.4 0.7,0 0,0.7 4.4,5.1 0,9.5 0.7,10.2 5.1,5.8 9.5,10.2 10.2,9.5 5.8,5.1 " />
      </symbol>
    </svg>
  </div>
</template>

<script lang="ts">
  import { HomeOutlined } from '@ant-design/icons-vue';
  import { Quit, WindowIsMaximised, WindowIsMinimised, WindowMaximise, WindowMinimise, WindowUnmaximise } from '@wailsapp/runtime';
  import { debounceTime, fromEvent, Subject, takeUntil } from 'rxjs';

  export default {
    name: 'Titlebar',
    props: {
      icon: { type: Boolean, default: true },
      theme: { type: String, default: 'dark' },
      title: { type: String },
      showMinimize: { type: Boolean, default: true },
      showMaximize: { type: Boolean, default: true },
      showClose: { type: Boolean, default: true },
      isMaximized: { type: Boolean, default: false }
    },
    emits: ['update:isMaximized', 'maximizeClick', 'minimizeClick', 'restoreDownClick', 'closing'],
    components: { HomeOutlined },
    setup(props, context) {
      const { emit } = context;
      const isMaximized_inner = ref(props.isMaximized);

      /**
       * 最大化点击事件
       *
       * @param
       */
      function onMaxIconClick() {
        isMaximized_inner.value = !isMaximized_inner.value;
        emit('update:isMaximized', isMaximized_inner.value);
        if (isMaximized_inner.value) {
          WindowMaximise();
          emit('maximizeClick', isMaximized_inner.value);
        } else {
          WindowUnmaximise();
          emit('restoreDownClick', isMaximized_inner.value);
        }
      }

      function onMinimizeClick() {
        WindowMinimise();
        emit('minimizeClick');
      }

      function onCloseClick() {
        const evt = { returnValue: true };
        emit('closing', evt);
        if (evt.returnValue) {
          Quit();
        }
      }

      const destroy$ = new Subject<void>();
      onMounted(() => {
        fromEvent(window, 'resize')
          .pipe(takeUntil(destroy$), debounceTime(100))
          .subscribe(async () => {
            const v = await WindowIsMinimised();
            if (!v) {
              isMaximized_inner.value = await WindowIsMaximised();
            }
          });
      });

      onUnmounted(() => {
        destroy$.next();
        destroy$.complete();
      });

      return { emit, onMaxIconClick, onMinimizeClick, onCloseClick, isMaximized_inner };
    }
  };
</script>

<style scoped lang="less">
  .desk-title-bar {
    user-select: none;
    --wails-draggable: drag;
    -webkit-app-region: drag;
    app-region: drag;
    cursor: default;
    display: flex;
    align-items: center;
    width: 100%;
    height: 32px;
    visibility: visible;
    position: relative;
    left: 0;
    top: 0;
    right: 0;
    z-index: 999999;
    padding-left: 12px;

    .left-bar {
      -webkit-tap-highlight-color: rgba(0, 0, 0, 0);
      font-size: 16px;
      margin-right: 8px;
    }

    .controls {
      user-select: none;
      cursor: default;
      display: flex;
      height: 32px;

      .button {
        -webkit-user-select: none;
        user-select: none;
        --wails-draggable: no-drag;
        -webkit-app-region: no-drag;
        app-region: no-drag;
        cursor: default;
        width: 46px;
        height: 100%;
        line-height: 0;
        display: flex;
        justify-content: center;
        align-items: center;
        fill: currentColor;

        &.minimize {
          stroke: currentColor;
        }
      }

      .icon {
        width: 10px;
        height: 10px;
      }
    }

    .title {
      user-select: none;
      cursor: default;
      font-size: 13px;
      flex: 1 1 0;
    }

    // 亮色主题
    &.light {
      .left-bar {
        color: #000;
      }
      .controls {
        .button {
          color: #000;

          &.close {
            &:hover {
              transition: background-color 0.1s;
              background-color: #e81123;
              color: #fff;
            }

            &:active {
              background-color: #f1707a;
              color: #fff;
            }
          }

          &.maximize,
          &.minimize {
            &:hover {
              transition: background-color 0.1s;
              background-color: #e5e5e5;
            }

            &:active {
              background-color: #ccc;
            }
          }
        }
      }
    }

    // 黑色主题
    &.dark {
      background-color: #0078d7;
      .left-bar {
        color: #fff;
      }
      .title {
        color: #fff;
      }
      .controls {
        .button {
          color: #fff;
          &.close {
            &:hover {
              transition: background-color 0.1s;
              background-color: #e81123;
            }
            &:active {
              background-color: #f1707a;
            }
          }

          &.maximize,
          &.minimize {
            &:hover {
              transition: background-color 0.1s;
              background-color: rgba(255, 255, 255, 0.13);
            }
            &:active {
              background-color: rgba(255, 255, 255, 0.23);
            }
          }
        }
      }
    }
  }
</style>
