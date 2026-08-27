import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface SidebarState {
  expandedNodes: Record<string, boolean>;
  toggleNode: (nodeId: string) => void;
  setNodeExpanded: (nodeId: string, expanded: boolean) => void;
}

export const useSidebarStore = create<SidebarState>()(
  persist(
    (set) => ({
      expandedNodes: {},
      toggleNode: (nodeId) => set((state) => ({
        expandedNodes: {
          ...state.expandedNodes,
          [nodeId]: !state.expandedNodes[nodeId]
        }
      })),
      setNodeExpanded: (nodeId, expanded) => set((state) => ({
        expandedNodes: {
          ...state.expandedNodes,
          [nodeId]: expanded
        }
      })),
    }),
    {
      name: 'dboke-sidebar-storage',
    }
  )
);
