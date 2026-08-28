"use client";

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useSidebarStore } from '../lib/store';
import { databasesApi, Table } from '../lib/api/databases';

const ChevronRightIcon = ({ className = "" }: { className?: string }) => (
  <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className={className}>
    <polyline points="9 18 15 12 9 6" />
  </svg>
);

const DatabaseIcon = ({ className = "" }: { className?: string }) => (
  <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className={className}>
    <ellipse cx="12" cy="5" rx="9" ry="3" />
    <path d="M3 5V19A9 3 0 0 0 21 19V5" />
    <path d="M3 12A9 3 0 0 0 21 12" />
  </svg>
);

const FolderIcon = ({ className = "" }: { className?: string }) => (
  <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className={className}>
    <path d="M4 20h16a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.93a2 2 0 0 1-1.66-.9l-.82-1.2A2 2 0 0 0 7.93 3H4a2 2 0 0 0-2 2v13c0 1.1.9 2 2 2Z" />
  </svg>
);

const TableIcon = ({ className = "" }: { className?: string }) => (
  <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className={className}>
    <rect width="18" height="18" x="3" y="3" rx="2" />
    <path d="M3 9h18" />
    <path d="M9 21V9" />
  </svg>
);

// Generic TreeNode Component
function TreeNode({ 
  id, 
  label, 
  icon: Icon, 
  children,
  onClick,
  href,
  level = 0,
  onNavigate
}: { 
  id: string, 
  label: string, 
  icon: any, 
  children?: React.ReactNode,
  onClick?: () => void,
  href?: string,
  level?: number,
  onNavigate?: () => void
}) {
  const { expandedNodes, toggleNode } = useSidebarStore();
  const isExpanded = !!expandedNodes[id];
  const hasChildren = !!children;

  const router = useRouter();

  const handleToggle = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (hasChildren) {
      toggleNode(id);
    }
    if (href) {
      router.push(href);
      if (onNavigate) onNavigate();
    }
    if (onClick) onClick();
  };

  const content = (
    <div 
      className="flex items-center gap-1.5 py-1 px-2 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-200/50 dark:hover:bg-gray-800/50 rounded-md cursor-pointer select-none transition-colors"
      style={{ paddingLeft: `${level * 12 + 8}px` }}
      onClick={handleToggle}
    >
      <div className="w-4 flex items-center justify-center">
        {hasChildren && (
          <ChevronRightIcon className={`text-gray-400 transition-transform ${isExpanded ? 'rotate-90' : ''}`} />
        )}
      </div>
      <Icon className={label === 'Schemas' || label === 'Columns' || label === 'Tables' ? "text-yellow-500" : "text-blue-500"} />
      <span className="truncate">{label}</span>
    </div>
  );

  return (
    <div>
      {content}
      
      {isExpanded && hasChildren && (
        <div className="flex flex-col">
          {children}
        </div>
      )}
    </div>
  );
}

// Table Node with dynamic columns
function TableNode({ dbName, tableName, level, onNavigate }: { dbName: string, tableName: string, level: number, onNavigate?: () => void }) {
  const { expandedNodes } = useSidebarStore();
  const isColumnsExpanded = !!expandedNodes[`${dbName}-${tableName}-columns`];
  const [columns, setColumns] = useState<any[]>([]);
  const [loadingColumns, setLoadingColumns] = useState(false);

  useEffect(() => {
    if (isColumnsExpanded && columns.length === 0 && !loadingColumns) {
      setLoadingColumns(true);
      databasesApi.getTableSchema(dbName, tableName)
        .then(res => {
          setColumns(res.columns || []);
        })
        .catch(console.error)
        .finally(() => setLoadingColumns(false));
    }
  }, [isColumnsExpanded, dbName, tableName, columns.length, loadingColumns]);

  return (
    <TreeNode id={`${dbName}-table-${tableName}`} label={tableName} icon={TableIcon} level={level} onNavigate={onNavigate}>
      <TreeNode id={`${dbName}-${tableName}-columns`} label="Columns" icon={FolderIcon} level={level + 1} onNavigate={onNavigate}>
        {columns.length > 0 ? (
          columns.map(c => (
            <TreeNode 
              key={c.name} 
              id={`${dbName}-${tableName}-col-${c.name}`} 
              label={`${c.name} (${c.type})`} 
              icon={FolderIcon} 
              level={level + 2} 
              onNavigate={onNavigate} 
            />
          ))
        ) : (
          isColumnsExpanded ? (
            <div className="py-1 text-xs text-gray-500 italic" style={{ paddingLeft: `${(level + 2) * 12 + 8}px` }}>
              {loadingColumns ? 'Loading columns...' : 'No columns found'}
            </div>
          ) : null
        )}
      </TreeNode>
      <TreeNode id={`${dbName}-${tableName}-constraints`} label="Constraints" icon={FolderIcon} level={level + 1} onNavigate={onNavigate} />
      <TreeNode id={`${dbName}-${tableName}-fks`} label="Foreign Keys" icon={FolderIcon} level={level + 1} onNavigate={onNavigate} />
      <TreeNode id={`${dbName}-${tableName}-idx`} label="Indexes" icon={FolderIcon} level={level + 1} onNavigate={onNavigate} />
      <TreeNode id={`${dbName}-${tableName}-dep`} label="Dependencies" icon={FolderIcon} level={level + 1} onNavigate={onNavigate} />
    </TreeNode>
  );
}

// Database Node that fetches tables when expanded
function DatabaseNode({ dbName, level, onNavigate }: { dbName: string, level: number, onNavigate?: () => void }) {
  const { expandedNodes } = useSidebarStore();
  const isExpanded = !!expandedNodes[`db-${dbName}`];
  const isTablesExpanded = !!expandedNodes[`${dbName}-tables`];
  const [tables, setTables] = useState<Table[]>([]);
  
  useEffect(() => {
    if (isTablesExpanded && tables.length === 0) {
      databasesApi.getTables(dbName).then(res => setTables(res.tables || [])).catch(console.error);
    }
  }, [isTablesExpanded, dbName, tables.length]);

  return (
    <TreeNode id={`db-${dbName}`} label={dbName} icon={DatabaseIcon} level={level} href={`/databases?id=${dbName}`} onNavigate={onNavigate}>
      <TreeNode id={`${dbName}-schemas`} label="Schemas" icon={FolderIcon} level={level + 1} onNavigate={onNavigate}>
        <TreeNode id={`${dbName}-public`} label="public" icon={FolderIcon} level={level + 2} onNavigate={onNavigate}>
          <TreeNode id={`${dbName}-tables`} label="Tables" icon={FolderIcon} level={level + 3} href={`/databases?id=${dbName}`} onNavigate={onNavigate}>
            {tables.length > 0 ? (
              tables.map(t => (
                <TableNode key={t.name} dbName={dbName} tableName={t.name} level={level + 4} onNavigate={onNavigate} />
              ))
            ) : (
              <div className="py-1 text-xs text-gray-500 italic" style={{ paddingLeft: `${(level + 4) * 12 + 8}px` }}>Loading tables...</div>
            )}
          </TreeNode>
          <TreeNode id={`${dbName}-views`} label="Views" icon={FolderIcon} level={level + 3} onNavigate={onNavigate} />
        </TreeNode>
      </TreeNode>
      <TreeNode id={`${dbName}-event-triggers`} label="Event Triggers" icon={FolderIcon} level={level + 1} onNavigate={onNavigate} />
      <TreeNode id={`${dbName}-extensions`} label="Extensions" icon={FolderIcon} level={level + 1} onNavigate={onNavigate} />
      <TreeNode id={`${dbName}-roles`} label="Roles" icon={FolderIcon} level={level + 1} onNavigate={onNavigate} />
    </TreeNode>
  );
}

export function SidebarTree({ databases, onNavigate }: { databases: string[], onNavigate?: () => void }) {
  return (
    <div className="py-2">
      <TreeNode id="root-databases" label="Databases" icon={FolderIcon} level={0} onNavigate={onNavigate}>
        {databases.map(dbName => (
          <DatabaseNode key={dbName} dbName={dbName} level={1} onNavigate={onNavigate} />
        ))}
      </TreeNode>
    </div>
  );
}
