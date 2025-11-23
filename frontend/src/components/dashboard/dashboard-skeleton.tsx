import { Card } from '../ui/card';

export function DashboardSkeleton() {
    return (
        <div className="min-h-screen bg-gray-50">
            {/* Header Skeleton */}
            <div className="bg-white border-b h-16 mb-8">
                <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-full flex justify-between items-center">
                    <div className="flex items-center space-x-4">
                        <div className="h-10 w-10 bg-gray-200 rounded-xl animate-pulse"></div>
                        <div className="space-y-2">
                            <div className="h-4 w-32 bg-gray-200 rounded animate-pulse"></div>
                            <div className="h-3 w-24 bg-gray-200 rounded animate-pulse"></div>
                        </div>
                    </div>
                    <div className="flex items-center space-x-4">
                        <div className="h-8 w-20 bg-gray-200 rounded animate-pulse"></div>
                        <div className="h-8 w-20 bg-gray-200 rounded animate-pulse"></div>
                    </div>
                </div>
            </div>

            <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pb-12">
                <div className="space-y-8">
                    {/* Quick Stats Skeleton */}
                    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
                        {[1, 2, 3, 4].map((i) => (
                            <Card key={i} className="p-6 border-none shadow-sm">
                                <div className="flex justify-between mb-4">
                                    <div className="h-10 w-10 bg-gray-200 rounded-xl animate-pulse"></div>
                                    <div className="h-4 w-16 bg-gray-200 rounded-full animate-pulse"></div>
                                </div>
                                <div className="space-y-2">
                                    <div className="h-8 w-12 bg-gray-200 rounded animate-pulse"></div>
                                    <div className="h-3 w-24 bg-gray-200 rounded animate-pulse"></div>
                                </div>
                            </Card>
                        ))}
                    </div>

                    {/* Owner Actions Skeleton */}
                    <Card className="p-6 border-none shadow-sm">
                        <div className="flex items-center mb-6">
                            <div className="h-10 w-10 bg-gray-200 rounded-xl mr-4 animate-pulse"></div>
                            <div className="space-y-2">
                                <div className="h-5 w-32 bg-gray-200 rounded animate-pulse"></div>
                                <div className="h-3 w-48 bg-gray-200 rounded animate-pulse"></div>
                            </div>
                        </div>
                        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
                            {[1, 2, 3].map((i) => (
                                <div key={i} className="h-32 bg-gray-100 rounded-lg animate-pulse"></div>
                            ))}
                        </div>
                    </Card>

                    <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
                        {/* Account Overview Skeleton */}
                        <div className="lg:col-span-1">
                            <Card className="p-6 border-none shadow-sm h-full">
                                <div className="flex justify-between mb-6">
                                    <div className="h-6 w-32 bg-gray-200 rounded animate-pulse"></div>
                                    <div className="h-8 w-20 bg-gray-200 rounded animate-pulse"></div>
                                </div>
                                <div className="space-y-6">
                                    {[1, 2, 3, 4].map((i) => (
                                        <div key={i} className="flex items-center">
                                            <div className="h-8 w-8 bg-gray-200 rounded-full mr-4 animate-pulse"></div>
                                            <div className="space-y-2 flex-1">
                                                <div className="h-3 w-20 bg-gray-200 rounded animate-pulse"></div>
                                                <div className="h-4 w-full bg-gray-200 rounded animate-pulse"></div>
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            </Card>
                        </div>

                        {/* Organizations List Skeleton */}
                        <div className="lg:col-span-2">
                            <Card className="p-6 border-none shadow-sm h-full">
                                <div className="flex justify-between mb-6">
                                    <div className="h-6 w-40 bg-gray-200 rounded animate-pulse"></div>
                                    <div className="h-8 w-24 bg-gray-200 rounded animate-pulse"></div>
                                </div>
                                <div className="space-y-4">
                                    {[1, 2, 3].map((i) => (
                                        <div key={i} className="h-24 bg-gray-100 rounded-xl animate-pulse"></div>
                                    ))}
                                </div>
                            </Card>
                        </div>
                    </div>
                </div>
            </main>
        </div>
    );
}
