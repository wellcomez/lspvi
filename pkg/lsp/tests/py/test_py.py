class Person:
    def __init__(self, name, age, city):
        self.name = name
        self.age = age
        self.city = city

    def greet(self):
        print(f"Hello, my name is {self.name} and I am {self.age} years old. I live in {self.city}.")

    def celebrate_birthday(self):
        self.age += 1
        print(f"Happy birthday! Now I am {self.age} years old.")


class Student(Person):
    def __init__(self, name, age, city, school):
        super().__init__(name, age, city)
        self.school = school

    def study(self):
        print(f"I am studying at {self.school}.")

# 创建一个 Student 对象
student1 = Student("Bob", 20, "San Francisco", "UC Berkeley")

# 调用 greet 方法
student1.greet()

# 调用 study 方法
student1.study()

# 庆祝生日
student1.celebrate_birthday()

# 再次调用 greet 方法
student1.greet()