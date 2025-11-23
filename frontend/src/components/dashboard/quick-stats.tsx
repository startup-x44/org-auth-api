import { Card } from '../ui/card';
import { User, Building2, Crown, Calendar } from 'lucide-react';

interface QuickStatsProps {
    isActive: boolean;
    organizationsCount: number;
    ownedOrganizationsCount: number;
    daysMember: number;
}

export function QuickStats({
    isActive,
    organizationsCount,
    ownedOrganizationsCount,
    daysMember,
}: QuickStatsProps) {
    const stats = [
        {
            label: 'Active Status',
            value: isActive ? 'Active' : 'Inactive',
            icon: User,
            color: isActive ? 'text-success' : 'text-destructive',
            bgColor: isActive ? 'bg-success/10' : 'bg-destructive/10',
            subtext: 'Account Status'
        },
        {
            label: 'Organizations',
            value: organizationsCount.toString(),
            icon: Building2,
            color: 'text-primary',
            bgColor: 'bg-primary/10',
            subtext: 'Total Memberships'
        },
        {
            label: 'Owned Orgs',
            value: ownedOrganizationsCount.toString(),
            icon: Crown,
            color: 'text-accent-foreground',
            bgColor: 'bg-accent',
            subtext: 'Organizations Owned'
        },
        {
            label: 'Days Member',
            value: daysMember.toString(),
            icon: Calendar,
            color: 'text-secondary-foreground',
            bgColor: 'bg-secondary/10',
            subtext: 'Since Joining'
        }
    ];

    return (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 sm:gap-6">
            {stats.map((stat, index) => (
                <Card key={index} className="p-6 hover:shadow-lg transition-all duration-300 border-none shadow-sm bg-card">
                    <div className="flex items-center justify-between mb-4">
                        <div className={`p-3 rounded-xl ${stat.bgColor}`}>
                            <stat.icon className={`h-6 w-6 ${stat.color}`} />
                        </div>
                        <span className={`text-xs font-semibold px-2.5 py-0.5 rounded-full ${stat.bgColor} ${stat.color}`}>
                            {stat.label}
                        </span>
                    </div>
                    <div>
                        <div className={`text-2xl font-bold ${stat.color}`}>
                            {stat.value}
                        </div>
                        <div className="text-sm text-muted-foreground font-medium mt-1">{stat.subtext}</div>
                    </div>
                </Card>
            ))}
        </div>
    );
}
