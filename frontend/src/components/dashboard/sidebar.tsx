'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import {
    LayoutDashboard,
    User,
    Building2,
    Settings,
    Shield,
    ChevronLeft,
    ChevronRight,
    Key,
    Lock,
    BookOpen
} from 'lucide-react';
import { useState } from 'react';

interface SidebarProps {
    hasOwnerPrivileges: boolean;
    organizationName?: string;
    isOpen?: boolean;
    onClose?: () => void;
    collapsed?: boolean;
    onCollapsedChange?: (collapsed: boolean) => void;
}

export function Sidebar({ hasOwnerPrivileges, organizationName = 'BlockSure', isOpen = false, onClose, collapsed = false, onCollapsedChange }: SidebarProps) {
    const pathname = usePathname();
    
    const handleToggleCollapse = () => {
        const newCollapsed = !collapsed;
        onCollapsedChange?.(newCollapsed);
    };

    const menuItems = [
        {
            name: 'Dashboard',
            href: '/user/dashboard',
            icon: LayoutDashboard,
            description: 'Overview and stats'
        },
        {
            name: 'Profile',
            href: '/user/profile',
            icon: User,
            description: 'Your account details'
        }
    ];

    // Only show Members, Settings, Roles & Permissions, OAuth2, and API Keys for owners
    if (hasOwnerPrivileges) {
        menuItems.push({
            name: 'Members',
            href: '/user/members',
            icon: Building2,
            description: 'Manage team members'
        });

        menuItems.push({
            name: 'Settings',
            href: '/user/settings',
            icon: Settings,
            description: 'Organization settings'
        });

        menuItems.push({
            name: 'Roles & Permissions',
            href: '/user/roles',
            icon: Shield,
            description: 'Access control'
        });

        menuItems.push({
            name: 'OAuth2 Apps',
            href: '/user/oauth2',
            icon: Lock,
            description: 'OAuth2 applications'
        });

        menuItems.push({
            name: 'API Keys',
            href: '/user/api-keys',
            icon: Key,
            description: 'API key management'
        });

        menuItems.push({
            name: 'Developer Docs',
            href: '/user/developer-docs',
            icon: BookOpen,
            description: 'Integration guides'
        });
    }

    // Mobile sidebar classes
    const mobileClasses = `transform lg:transform-none lg:opacity-100 transition-all duration-300 ease-in-out z-[60] fixed inset-y-0 left-0 bg-background border-r border-border lg:fixed lg:h-screen ${isOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'
        }`;

    // Desktop width classes
    const desktopWidthClass = collapsed ? 'lg:w-16' : 'lg:w-64';
    const widthClass = 'w-64'; // Always full width on mobile

    return (
        <>
            <aside className={`${mobileClasses} ${widthClass} ${desktopWidthClass}`}>
                <div className="flex flex-col h-full">
                    {/* Logo/Brand */}
                    <div className="h-16 flex items-center justify-between px-4 border-b border-border">
                        {(!collapsed || isOpen) && (
                            <div className="flex items-center space-x-2">
                                <div className="w-8 h-8 bg-primary rounded-lg flex items-center justify-center">
                                    <span className="text-primary-foreground font-bold text-sm">
                                        {organizationName.substring(0, 2).toUpperCase()}
                                    </span>
                                </div>
                                <span className="font-bold text-foreground">{organizationName}</span>
                            </div>
                        )}
                        {/* Desktop collapse button */}
                        <button
                            onClick={handleToggleCollapse}
                            className="hidden lg:block p-1.5 hover:bg-muted rounded-lg transition-colors"
                        >
                            {collapsed ? (
                                <ChevronRight className="h-5 w-5 text-muted-foreground" />
                            ) : (
                                <ChevronLeft className="h-5 w-5 text-muted-foreground" />
                            )}
                        </button>
                        {/* Mobile close button */}
                        <button
                            onClick={onClose}
                            className="lg:hidden p-1.5 hover:bg-muted rounded-lg transition-colors"
                        >
                            <ChevronLeft className="h-5 w-5 text-muted-foreground" />
                        </button>
                    </div>

                    {/* Navigation */}
                    <nav className="flex-1 overflow-y-auto py-4 px-2">
                        <div className="space-y-1">
                            {menuItems.map((item) => {
                                const isActive = pathname === item.href;
                                const Icon = item.icon;

                                return (
                                    <Link
                                        key={item.href}
                                        href={item.href}
                                        onClick={() => onClose?.()} // Close sidebar on mobile when link clicked
                                        className={`flex items-center px-3 py-2.5 rounded-lg transition-all duration-200 group ${isActive
                                            ? 'bg-primary/10 text-primary'
                                            : 'text-foreground hover:bg-muted hover:text-foreground'
                                            }`}
                                    >
                                        <Icon className={`h-5 w-5 flex-shrink-0 ${isActive ? 'text-primary' : 'text-muted-foreground group-hover:text-foreground'
                                            }`} />
                                        {(!collapsed || isOpen) && (
                                            <div className="ml-3 flex-1">
                                                <div className={`text-sm font-medium ${isActive ? 'text-primary' : 'text-foreground'
                                                    }`}>
                                                    {item.name}
                                                </div>
                                                <div className="text-xs text-muted-foreground mt-0.5">
                                                    {item.description}
                                                </div>
                                            </div>
                                        )}
                                    </Link>
                                );
                            })}
                        </div>
                    </nav>

                    {/* Footer */}
                    {(!collapsed || isOpen) && (
                        <div className="p-4 border-t border-border">
                            <div className="text-xs text-muted-foreground text-center">
                                © 2025 NILOAUTH
                            </div>
                        </div>
                    )}
                </div>
            </aside>
        </>
    );
}
